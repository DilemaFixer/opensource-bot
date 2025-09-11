package githubapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	GitHubRepoAPI struct {
		GitHubAPI
		ID               int64     `json:"id"`
		NodeID           string    `json:"node_id"`
		Name             string    `json:"name"`
		FullName         string    `json:"full_name"`
		Private          bool      `json:"private"`
		Owner            RepoOwner `json:"owner"`
		HTMLURL          string    `json:"html_url"`
		Description      *string   `json:"description"`
		Fork             bool      `json:"fork"`
		URL              string    `json:"url"`
		CreatedAt        string    `json:"created_at"`
		UpdatedAt        string    `json:"updated_at"`
		PushedAt         string    `json:"pushed_at"`
		GitURL           string    `json:"git_url"`
		SSHURL           string    `json:"ssh_url"`
		CloneURL         string    `json:"clone_url"`
		SvnURL           string    `json:"svn_url"`
		Homepage         *string   `json:"homepage"`
		Size             int       `json:"size"`
		StargazersCount  int       `json:"stargazers_count"`
		WatchersCount    int       `json:"watchers_count"`
		Language         *string   `json:"language"`
		HasIssues        bool      `json:"has_issues"`
		HasProjects      bool      `json:"has_projects"`
		HasDownloads     bool      `json:"has_downloads"`
		HasWiki          bool      `json:"has_wiki"`
		HasPages         bool      `json:"has_pages"`
		HasDiscussions   bool      `json:"has_discussions"`
		ForksCount       int       `json:"forks_count"`
		MirrorURL        *string   `json:"mirror_url"`
		Archived         bool      `json:"archived"`
		Disabled         bool      `json:"disabled"`
		OpenIssuesCount  int       `json:"open_issues_count"`
		License          *License  `json:"license"`
		AllowForking     bool      `json:"allow_forking"`
		IsTemplate       bool      `json:"is_template"`
		WebCommitSignoff bool      `json:"web_commit_signoff_required"`
		Topics           []any     `json:"topics"`
		Visibility       string    `json:"visibility"`
		Forks            int       `json:"forks"`
		OpenIssues       int       `json:"open_issues"`
		Watchers         int       `json:"watchers"`
		DefaultBranch    string    `json:"default_branch"`
		NetworkCount     int       `json:"network_count"`
		SubscribersCount int       `json:"subscribers_count"`
		TempCloneToken   *string   `json:"temp_clone_token"`
	}

	RepoOwner struct {
		Login             string `json:"login"`
		ID                int64  `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		UserViewType      string `json:"user_view_type"`
		SiteAdmin         bool   `json:"site_admin"`
	}
	License struct {
		Key    string  `json:"key"`
		Name   string  `json:"name"`
		SpdxID string  `json:"spdx_id"`
		URL    *string `json:"url"`
		NodeID string  `json:"node_id"`
	}

	CreateContentResp struct {
		Content struct {
			Name    string `json:"name"`
			Path    string `json:"path"`
			SHA     string `json:"sha"`
			HTMLURL string `json:"html_url"`
		} `json:"content"`
		Commit struct {
			SHA     string `json:"sha"`
			Message string `json:"message"`
			HTMLURL string `json:"html_url"`
		} `json:"commit"`
	}
)

var ghTimeLayouts = []string{
	time.RFC3339,     // 2025-05-15T17:12:53Z
	time.RFC3339Nano, // 2025-05-15T17:12:53.123Z
}

func parseGitHubTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	var lastErr error
	for _, layout := range ghTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %q: %v", s, lastErr)
}

func formatTime(t time.Time, layout string, loc *time.Location) string {
	if loc != nil {
		t = t.In(loc)
	}
	if layout == "" {
		layout = "2006-01-02 15:04:05 MST"
	}
	return t.Format(layout)
}

func humanizeInt(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000.0)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000.0)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func humanizeSince(t time.Time, now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	d := now.Sub(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%d months ago", months)
	}
	years := months / 12
	return fmt.Sprintf("%d years ago", years)
}

func (r *GitHubRepoAPI) CreateAt() (time.Time, error) {
	return parseGitHubTime(r.CreatedAt)
}

func (r *GitHubRepoAPI) GetCreatedAt() (time.Time, error) {
	return parseGitHubTime(r.CreatedAt)
}

func (r *GitHubRepoAPI) GetUpdatedAt() (time.Time, error) {
	return parseGitHubTime(r.UpdatedAt)
}

func (r *GitHubRepoAPI) GetPushedAt() (time.Time, error) {
	return parseGitHubTime(r.PushedAt)
}

func (r *GitHubRepoAPI) FormatCreatedAt(layout string, loc *time.Location) (string, error) {
	t, err := r.GetCreatedAt()
	if err != nil {
		return "", err
	}
	return formatTime(t, layout, loc), nil
}

func (r *GitHubRepoAPI) FormatUpdatedAt(layout string, loc *time.Location) (string, error) {
	t, err := r.GetUpdatedAt()
	if err != nil {
		return "", err
	}
	return formatTime(t, layout, loc), nil
}

func (r *GitHubRepoAPI) FormatPushedAt(layout string, loc *time.Location) (string, error) {
	t, err := r.GetPushedAt()
	if err != nil {
		return "", err
	}
	return formatTime(t, layout, loc), nil
}

func (r *GitHubRepoAPI) UpdatedAgo(now time.Time) (string, error) {
	t, err := r.GetUpdatedAt()
	if err != nil {
		return "", err
	}
	return humanizeSince(t, now), nil
}

func (r *GitHubRepoAPI) GetHomepage() string {
	if r.Homepage != nil {
		return *r.Homepage
	}
	return ""
}

func (r *GitHubRepoAPI) GetLanguage() string {
	if r.Language != nil {
		return *r.Language
	}
	return ""
}

func (r *GitHubRepoAPI) HasLicense() bool {
	return r.License != nil && r.License.Name != ""
}

func (r *GitHubRepoAPI) GetLicenseName() string {
	if r.License != nil {
		return r.License.Name
	}
	return ""
}

func (r *GitHubRepoAPI) GetTopics() []string {
	if r.Topics == nil {
		return nil
	}
	out := make([]string, 0, len(r.Topics))
	for _, v := range r.Topics {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func (r *GitHubRepoAPI) Stars() int            { return r.StargazersCount }
func (r *GitHubRepoAPI) StarsHuman() string    { return humanizeInt(r.StargazersCount) }
func (r *GitHubRepoAPI) ForksHuman() string    { return humanizeInt(r.ForksCount) }
func (r *GitHubRepoAPI) WatchersHuman() string { return humanizeInt(r.Watchers) }

func (r *GitHubRepoAPI) PreferredCloneURL(ssh bool) string {
	if ssh && r.SSHURL != "" {
		return r.SSHURL
	}
	if r.CloneURL != "" {
		return r.CloneURL
	}
	if r.GitURL != "" {
		return r.GitURL
	}
	return r.HTMLURL
}

func (r *GitHubRepoAPI) IsPublic() bool {
	return !r.Private && strings.EqualFold(r.Visibility, "public")
}
func (r *GitHubRepoAPI) IsArchived() bool { return r.Archived }
func (r *GitHubRepoAPI) IsDisabled() bool { return r.Disabled }

func (r *GitHubRepoAPI) GetRepoTags() ([]string, error) {
	req, err := r.newReq(context.Background(), "GET",
		fmt.Sprintf("/repos/%s/%s/tags", url.PathEscape(r.Owner.Login), url.PathEscape(r.Name)), nil)
	if err != nil {
		return nil, err
	}

	var raw []struct {
		Name string `json:"name"`
	}
	if err := r.doJSON(req, &raw); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(raw))
	for _, t := range raw {
		names = append(names, t.Name)
	}
	return names, nil
}

func (r *GitHubRepoAPI) UploadMdFile(filename, content string) error {
	_, err := r.UploadMdFileWithToken(context.Background(), filename, content, "", "")
	return err
}

func (r *GitHubRepoAPI) UploadMdFileWithToken(
	ctx context.Context,
	filename string,
	content string,
	commitMsg string,
	branch string,
) (*CreateContentResp, error) {
	fn := strings.TrimSpace(filename)
	if fn == "" {
		return nil, fmt.Errorf("filename is empty")
	}
	if !strings.HasSuffix(strings.ToLower(fn), ".md") {
		fn += ".md"
	}
	fn = strings.TrimLeft(fn, "/")

	if commitMsg == "" {
		commitMsg = "chore: add " + fn
	}
	if branch == "" {
		branch = r.DefaultBranch
	}

	body := map[string]any{
		"message": commitMsg,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
	}
	if branch != "" {
		body["branch"] = branch
	}

	b, _ := json.Marshal(body)

	reqPath := fmt.Sprintf("/repos/%s/%s/contents/%s",
		url.PathEscape(r.Owner.Login),
		url.PathEscape(r.Name),
		pathEscapeSegments(fn),
	)

	req, err := r.newReq(ctx, http.MethodPut, reqPath, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if r.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.accessToken)
	}

	var out CreateContentResp
	if err := r.doJSON(req, &out); err != nil {
		if he, ok := err.(*HTTPError); ok && he.StatusCode == http.StatusUnprocessableEntity {
			return nil, fmt.Errorf("file %q already exists on branch %q: %s", fn, branch, he.Body)
		}
		return nil, err
	}
	return &out, nil
}

func pathEscapeSegments(p string) string {
	segs := strings.Split(p, "/")
	for i := range segs {
		segs[i] = url.PathEscape(segs[i])
	}
	return strings.Join(segs, "/")
}
