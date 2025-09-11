package githubapi

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type (
	GitHubProfileAPI struct {
		GitHubAPI
		Login             string    `json:"login"`
		ID                int64     `json:"id"`
		NodeID            string    `json:"node_id"`
		AvatarURL         string    `json:"avatar_url"`
		GravatarID        string    `json:"gravatar_id"`
		URL               string    `json:"url"`
		HTMLURL           string    `json:"htmlurl"`
		FollowersURL      string    `json:"followers_url"`
		FollowingURL      string    `json:"following_url"`
		GistsURL          string    `json:"gists_url"`
		StarredURL        string    `json:"starred_url"`
		SubscriptionsURL  string    `json:"subscriptions_url"`
		OrganizationsURL  string    `json:"organizations_url"`
		ReposURL          string    `json:"repos_url"`
		EventsURL         string    `json:"events_url"`
		ReceivedEventsURL string    `json:"received_events_url"`
		Type              string    `json:"type"`
		UserViewType      string    `json:"user_view_type"`
		SiteAdmin         bool      `json:"site_admin"`
		Name              string    `json:"name"`
		Company           string    `json:"company"`
		Blog              string    `json:"blog"`
		Location          string    `json:"location"`
		Email             *string   `json:"email"`
		Hireable          *bool     `json:"hireable"`
		Bio               *string   `json:"bio"`
		TwitterUsername   *string   `json:"twitter_username"`
		PublicRepos       int       `json:"public_repos"`
		PublicGists       int       `json:"public_gists"`
		Followers         int       `json:"followers"`
		Following         int       `json:"following"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
	}
)

func (p *GitHubProfileAPI) GetPublicRepos() ([]GitHubRepoAPI, error) {
	req, err := p.newReq(context.Background(), "GET", fmt.Sprintf("/users/%s/repos", p.Login), nil)
	if err != nil {
		return nil, err
	}

	var repos []GitHubRepoAPI
	if jerror := p.doJSON(req, &repos); jerror != nil {
		return nil, jerror
	}

	return repos, nil
}

func (p *GitHubProfileAPI) IsRepoExist(name string) (bool, error) {
	req, err := p.newReq(context.Background(), "GET", fmt.Sprintf("/repos/%s/%s", p.Login, name), nil)
	if err != nil {
		return false, err
	}

	if jerror := p.doJSON(req, &GitHubRepoAPI{}); jerror != nil {
		if he, ok := jerror.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (p *GitHubProfileAPI) GetRepoIfExist(name string) (*GitHubRepoAPI, error) {
	req, err := p.newReq(context.Background(), "GET", fmt.Sprintf("/repos/%s/%s", p.Login, name), nil)
	if err != nil {
		return nil, err
	}

	repo := &GitHubRepoAPI{}
	if jerror := p.doJSON(req, repo); jerror != nil {
		if he, ok := jerror.(*HTTPError); ok && he.StatusCode == http.StatusNotFound {
			return nil, NewRepoNotFoundError(name)
		}
		return nil, err
	}

	return repo, nil
}
