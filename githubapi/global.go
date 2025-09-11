package githubapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	Doer interface {
		Do(req *http.Request) (*http.Response, error)
	}

	GitHubAPI struct {
		oAuth       OAuthApp
		baseURL     string
		userAgent   string
		timeout     time.Duration
		http        Doer
		accessToken string
	}

	OAuthApp struct {
		ClientID     string
		ClientSecret string
		RedirectURI  string
	}

	PKCE struct {
		Verifier  string
		Challenge string
		Method    string
	}
)

func NewPKCE() (*PKCE, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(buf)

	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	return &PKCE{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

func NewDefaultGitHubAPI() *GitHubAPI {
	return &GitHubAPI{
		baseURL:   "https://api.github.com",
		http:      &http.Client{Timeout: 10 * time.Second},
		userAgent: "githubapi/1.0",
		timeout:   10 * time.Second,
	}
}

func WithOptions(opts ...Option) *GitHubAPI {
	def := NewDefaultGitHubAPI()

	for _, o := range opts {
		o(def)
	}

	return def
}

func (c *GitHubAPI) CheckUserExists(ctx context.Context, username string) (bool, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return false, errors.New("username is empty")
	}

	req, err := c.newReq(ctx, http.MethodGet, "/users/"+url.PathEscape(username), nil)
	if err != nil {
		return false, err
	}

	if err := c.doJSON(req, nil); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *GitHubAPI) GetUserIfExists(ctx context.Context, username string) (*GitHubProfileAPI, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username is empty")
	}

	req, err := c.newReq(ctx, http.MethodGet, "/users/"+url.PathEscape(username), nil)
	if err != nil {
		return nil, err
	}

	profile := &GitHubProfileAPI{}
	if err := c.doJSON(req, profile); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return nil, NewProfileNotFoundError(username)
		}
		return nil, err
	}
	profile.applyFrom(c)
	return profile, nil
}

func (c *GitHubAPI) CheckUserExistsById(ctx context.Context, id int64) (bool, error) {
	req, err := c.newReq(ctx, http.MethodGet, "/user/"+url.PathEscape(strconv.FormatInt(id, 10)), nil)
	if err != nil {
		return false, err
	}

	if err := c.doJSON(req, nil); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *GitHubAPI) CheckIsRepoExistsById(ctx context.Context, id int64) (bool, error) {
	req, err := c.newReq(ctx, http.MethodGet, "/repositories/"+url.PathEscape(strconv.FormatInt(id, 10)), nil)
	if err != nil {
		return false, err
	}

	repo := &GitHubRepoAPI{}
	if err := c.doJSON(req, repo); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return false, nil
		}
	}
	return true, nil
}

func (c *GitHubAPI) GetRepoByIdIfExist(ctx context.Context, id int64) (*GitHubRepoAPI, error) {
	req, err := c.newReq(ctx, http.MethodGet, "/repositories/"+url.PathEscape(strconv.FormatInt(id, 10)), nil)
	if err != nil {
		return nil, err
	}

	repo := &GitHubRepoAPI{}
	if err := c.doJSON(req, repo); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return nil, NewRepoNotFoundError(fmt.Sprint(id))
		}
		return nil, err
	}

	return repo, nil
}

func (c *GitHubAPI) GetUserByIdIfExist(ctx context.Context, id int64) (*GitHubProfileAPI, error) {
	req, err := c.newReq(ctx, http.MethodGet, "/user/"+url.PathEscape(strconv.FormatInt(id, 10)), nil)
	if err != nil {
		return nil, err
	}

	profile := &GitHubProfileAPI{}
	if err := c.doJSON(req, profile); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return nil, NewProfileNotFoundError(strconv.FormatInt(id, 10))
		}
		return nil, err
	}
	profile.applyFrom(c)
	return profile, nil
}

// pkce — PKCE data. can be nil
func (c *GitHubAPI) AuthURL(username, state string, scopes []string, allowSignup bool, pkce *PKCE) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.oAuth.ClientID)
	if c.oAuth.RedirectURI != "" {
		v.Set("redirect_uri", c.oAuth.RedirectURI)
	}
	if state != "" {
		v.Set("state", state)
	}
	if len(scopes) > 0 {
		v.Set("scope", strings.Join(scopes, " "))
	}
	if username != "" {
		v.Set("login", username)
	}
	if !allowSignup {
		v.Set("allow_signup", "false")
	}
	if pkce != nil {
		v.Set("code_challenge", pkce.Challenge)
		v.Set("code_challenge_method", pkce.Method)
	}

	return "https://github.com/login/oauth/authorize?" + v.Encode(), nil
}

func (p *GitHubAPI) applyFrom(api *GitHubAPI) {
	p.http = api.http
	p.oAuth = api.oAuth
	p.baseURL = api.baseURL
	p.userAgent = api.userAgent
	p.timeout = api.timeout
}

func (c *GitHubAPI) newReq(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	return req, nil
}

func (c *GitHubAPI) doJSON(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(b)}
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}
