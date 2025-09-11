package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	tb "gopkg.in/telebot.v4"
)

// ====== MODELS ======
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type AuthSession struct {
	ChatID         int64
	State          string
	RequestedLogin string
}

// ====== GLOBALS ======
var (
	bot *tb.Bot

	authMu       sync.Mutex
	authSessions = make(map[string]*AuthSession)
)

// ====== MAIN ======
func main() {
	var err error
	bot, err = tb.NewBot(tb.Settings{
		Token:  TELEGRAM_BOT_TOKEN,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized as @%s", bot.Me.Username)

	// OAuth callback сервер
	go startWebServer()

	// Handlers
	bot.Handle("/start", func(c tb.Context) error {
		return c.Send("Привет! Отправь мне свой GitHub username для верификации владения аккаунтом.")
	})

	bot.Handle("/verify", func(c tb.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return c.Send("Использование: /verify <github_username>")
		}
		return handleUsernameInput(c, args[0])
	})

	// Любой текст = попытка принять username
	bot.Handle(tb.OnText, func(c tb.Context) error {
		text := strings.TrimSpace(c.Message().Text)
		// Игнорируем сообщения, начинающиеся с '/'
		if strings.HasPrefix(text, "/") || text == "" {
			return nil
		}
		return handleUsernameInput(c, text)
	})

	bot.Start()
}

// ====== TELEGRAM FLOW ======
func handleUsernameInput(c tb.Context, username string) error {
	username = strings.TrimSpace(username)

	exists, err := checkUserExists(username)
	if err != nil {
		return c.Send(fmt.Sprintf("⚠️ Ошибка при проверке: %v", err))
	}
	if !exists {
		return c.Send(fmt.Sprintf("❌ Пользователь @%s не найден на GitHub", username))
	}

	chatID := c.Chat().ID
	state := generateState(chatID)

	// сохраняем сессию
	authMu.Lock()
	authSessions[state] = &AuthSession{
		ChatID:         chatID,
		State:          state,
		RequestedLogin: username,
	}
	authMu.Unlock()

	// OAuth URL
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=user:email",
		GITHUB_CLIENT_ID,
		url.QueryEscape(REDIRECT_URI),
		state,
	)

	// Inline кнопка
	btn := tb.InlineButton{Text: "🔐 Подтвердить владение через GitHub", URL: authURL}
	markup := &tb.ReplyMarkup{InlineKeyboard: [][]tb.InlineButton{{btn}}}

	return c.Send(
		fmt.Sprintf("Для подтверждения владения аккаунтом @%s нажми на кнопку ниже:", username),
		markup,
	)
}

// ====== GITHUB ======
func checkUserExists(username string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "TelegramBot/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return resp.StatusCode == http.StatusOK, nil
}

func exchangeCodeForToken(code string) (token string, scopes []string, tokenType string, err error) {
	data := url.Values{}
	data.Set("client_id", GITHUB_CLIENT_ID)
	data.Set("client_secret", GITHUB_CLIENT_SECRET)
	data.Set("code", code)
	// если есть redirect_uri/state — добавь:
	// data.Set("redirect_uri", REDIRECT_URI)
	// data.Set("state", state)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", nil, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, "", err
	}

	var r struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`      // строка вида "user:email,read:user"
		TokenType   string `json:"token_type"` // обычно "bearer"
		Error       string `json:"error"`      // если было
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", nil, "", err
	}
	if r.Error != "" {
		if r.ErrorDesc != "" {
			return "", nil, "", fmt.Errorf("GitHub error: %s (%s)", r.Error, r.ErrorDesc)
		}
		return "", nil, "", fmt.Errorf("GitHub error: %s", r.Error)
	}

	// разберём scope в []string
	var scopeSlice []string
	if s := strings.TrimSpace(r.Scope); s != "" {
		for _, part := range strings.Split(s, ",") {
			p := strings.TrimSpace(part)
			if p != "" {
				scopeSlice = append(scopeSlice, p)
			}
		}
	}

	return r.AccessToken, scopeSlice, r.TokenType, nil
}

func getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ====== HTTP CALLBACK ======
func startWebServer() {
	http.HandleFunc("/callback", handleGitHubCallback)
	log.Println("Starting web server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start web server:", err)
	}
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}

	// достаём и удаляем сессию
	authMu.Lock()
	session, ok := authSessions[state]
	if ok {
		delete(authSessions, state)
	}
	authMu.Unlock()
	if !ok {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// меняем code на токен
	accessToken, scopeSlice, tokenType, err := exchangeCodeForToken(code)
	_ = scopeSlice
	_ = tokenType
	if err != nil {
		log.Printf("exchange error: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		_, _ = bot.Send(&tb.User{ID: session.ChatID}, "❌ Ошибка авторизации")
		return
	}

	// получаем пользователя
	user, err := getGitHubUser(accessToken)
	if err != nil {
		log.Printf("user info error: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		_, _ = bot.Send(&tb.User{ID: session.ChatID}, "❌ Не удалось получить информацию о пользователе")
		return
	}

	// сверяем логин
	if strings.ToLower(user.Login) != strings.ToLower(session.RequestedLogin) {
		_, _ = bot.Send(&tb.User{ID: session.ChatID},
			fmt.Sprintf("❌ Ошибка верификации!\n\nЗапрашивался: @%s\nАвторизован: @%s\n\nПожалуйста, авторизуйтесь под правильным аккаунтом.",
				session.RequestedLogin, user.Login))

		fmt.Fprintf(w, `
<html>
<head><title>Ошибка верификации</title></head>
<body>
<h1>❌ Ошибка верификации</h1>
<p>Авторизован неправильный аккаунт. Закройте страницу и попробуйте снова.</p>
</body>
</html>`)
		return
	}

	// успех
	_, _ = bot.Send(&tb.User{ID: session.ChatID},
		fmt.Sprintf("✅ Владение аккаунтом подтверждено!\n\n👤 Имя: %s\n🔗 GitHub: @%s\n📧 Email: %s\n🆔 ID: %d",
			emptyIf(user.Name, "—"), user.Login, emptyIf(user.Email, "—"), user.ID))

	log.Printf("User verified: %s (ID: %d, Chat: %d)", user.Login, user.ID, session.ChatID)

	fmt.Fprintf(w, `
<html>
<head><title>Верификация успешна</title></head>
<body style="font-family: Arial, sans-serif; text-align: center; margin-top: 50px;">
<h1 style="color: green;">✅ Владение аккаунтом подтверждено!</h1>
<p>Аккаунт <strong>@%s</strong> успешно верифицирован.</p>
<p>Можете закрыть эту страницу и вернуться в Telegram.</p>
</body>
</html>`, user.Login)
}

// ====== UTILS ======
func generateState(chatID int64) string {
	return fmt.Sprintf("%d_%d", chatID, time.Now().UnixNano())
}

func emptyIf(s, repl string) string {
	if strings.TrimSpace(s) == "" {
		return repl
	}
	return s
}
