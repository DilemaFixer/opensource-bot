package githubapi

import (
	"fmt"
)

type (
	HTTPError struct {
		StatusCode int
		Body       string
	}

	ProfileNotFoundError struct {
		Profile string
	}

	RepoNotFoundError struct {
		Repo string
	}
)

func (e *HTTPError) Error() string {
	return fmt.Sprintf("github: http %d: %s", e.StatusCode, e.Body)
}
func NewProfileNotFoundError(profile string) error { return &ProfileNotFoundError{profile} }
func (e *ProfileNotFoundError) Error() string      { return fmt.Sprintf("profile not found: %s", e.Profile) }

func (e *RepoNotFoundError) Error() string   { return fmt.Sprintf("repo not found: %s", e.Repo) }
func NewRepoNotFoundError(repo string) error { return &RepoNotFoundError{repo} }
