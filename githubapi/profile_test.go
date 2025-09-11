package githubapi

import (
	"context"
	"testing"
)

const (
	GitHubUserId        int64  = 109586143
	ExistingRepoName    string = "Annuum"
	NonExistingRepoName string = "labyba"
)

func MakeBaseUserAPI() (*GitHubProfileAPI, error) {
	gapi := NewDefaultGitHubAPI()
	papi, err := gapi.GetUserByIdIfExist(context.Background(), GitHubUserId)
	if err != nil {
		return nil, err
	}
	return papi, nil
}

func TestGetPublicRepos_WithValidData_MustReturnReposSlice(t *testing.T) {
	papi, err := MakeBaseUserAPI()
	if err != nil {
		t.Fatal(err)
	}

	repos, err := papi.GetPublicRepos()
	if err != nil {
		t.Fatal(err)
	}

	_ = repos
}

func TestIsRepoExist_WithExistingRepoName_ReturnsTrue(t *testing.T) {
	papi, err := MakeBaseUserAPI()
	if err != nil {
		t.Fatal(err)
	}

	isExist, err := papi.IsRepoExist(ExistingRepoName)
	if err != nil {
		t.Fatal(err)
	}

	if !isExist {
		t.Fatalf("Repo %s does not exist , but must", ExistingRepoName)
	}
}

func TestIsRepoExist_WithNonExistingRepoName_ReturnsFalse(t *testing.T) {
	papi, err := MakeBaseUserAPI()
	if err != nil {
		t.Fatal(err)
	}
	isExist, err := papi.IsRepoExist(NonExistingRepoName)
	if err != nil {
		t.Fatal(err)
	}

	if isExist {
		t.Fatalf("Repo %s exist , but must fail", NonExistingRepoName)
	}
}

func TestIsRepoExist_WithExistingRepoName_ReturnsFalse(t *testing.T) {
	gapi, err := MakeBaseUserAPI()
	if err != nil {
		t.Fatal(err)
	}

	papi, err := gapi.GetUserByIdIfExist(context.Background(), GitHubUserId)
	if err != nil {
		t.Fatal(err)
	}

	repo, err := papi.GetRepoIfExist(ExistingRepoName)
	if err != nil {
		t.Fatal(err)
	}

	if repo.Name != ExistingRepoName {
		t.Fatalf("Repo %s exist , but must fail", ExistingRepoName)
	}
}

func TestIsRepoExist_WithNonExistingRepoName_ReturnsTrue(t *testing.T) {
	gapi, g_err := MakeBaseUserAPI()
	if g_err != nil {
		t.Fatal(g_err)
	}

	papi, u_err := gapi.GetUserByIdIfExist(context.Background(), GitHubUserId)
	if u_err != nil {
		t.Fatal(u_err)
	}

	_, err := papi.GetRepoIfExist(NonExistingRepoName)
	if err != nil {
		if er := err.(*RepoNotFoundError); er.Repo != NonExistingRepoName {
			t.Fatalf("Repo %s exist , but must fail", NonExistingRepoName)
		}
	}
}
