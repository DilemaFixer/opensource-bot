package githubapi

import (
	"time"
)

type Option func(*GitHubAPI)

func WithHTTP(http Doer) Option {
	return func(api *GitHubAPI) {
		api.http = http
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(api *GitHubAPI) {
		api.timeout = timeout
	}
}

func WithUserAgent(userAgent string) Option {
	return func(api *GitHubAPI) {
		api.userAgent = userAgent
	}
}

func WithBaseURL(baseURL string) Option {
	return func(api *GitHubAPI) {
		api.baseURL = baseURL
	}
}

func WithOAuth(oAuth OAuthApp) Option {
	return func(api *GitHubAPI) {
		api.oAuth = oAuth
	}
}
