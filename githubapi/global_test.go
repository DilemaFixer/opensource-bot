package githubapi

import (
	"context"
	"testing"
)

const (
	ExistingUser                = "DilemaFixer"
	NonExistingUser             = "miiiiiiiiiaaaaaaaaaaayyyyyyyyy"
	ExistingUserId        int64 = 109586143
	ExistingRepoId        int64 = 984280187
	NonExistingRepoId     int64 = 0
	NonExistingUserId     int64 = 0
	NonExistingIdAsString       = "0"
)

func TestCheckUserExists_WithNonExistingUsername_MustReturnFalse(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	isExist, err := gapi.CheckUserExists(context.Background(), NonExistingUser)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if isExist {
		t.Fatalf("user %s should not exist", NonExistingUser)
	}
}

func TestCheckUserExists_WithExistingUsername_MustReturnTrue(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	isExist, err := gapi.CheckUserExists(context.Background(), ExistingUser)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !isExist {
		t.Fatalf("user %s should exist", ExistingUser)
	}
}

func TestGetUserIfExists_WithNonExistingUsername_MustReturnNotFoundErr(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	_, err := gapi.GetUserIfExists(context.Background(), NonExistingUser)
	if err != nil {
		if er := err.(*ProfileNotFoundError); er.Profile != NonExistingUser {
			t.Fatalf("got unexpected error: %s", err)
		}
	}
}

func TestGetUserIfExists_WithExistingUsername_MustReturnUserStruct(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	user, err := gapi.GetUserIfExists(context.Background(), ExistingUser)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if user == nil {
		t.Fatalf("got nil user")
	}
	if user.ID != ExistingUserId {
		t.Fatalf("got unexpected user ID %d", user.ID)
	}
}

func TestCheckUserExistsById_WithNonExistingId_MustReturnFalse(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	isExist, err := gapi.CheckUserExistsById(context.Background(), NonExistingUserId)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if isExist {
		t.Fatalf("user %d should not exist", NonExistingUserId)
	}
}

func TestCheckUserExistsById_WithExistId_MustReturnTrue(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	isExist, err := gapi.CheckUserExistsById(context.Background(), ExistingUserId)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !isExist {
		t.Fatalf("user %d should exist", ExistingUserId)
	}
}

func TestGetUserByIdIfExist_WithNonExistingId_MustReturnNotFoundErr(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	_, err := gapi.GetUserByIdIfExist(context.Background(), NonExistingUserId)
	if err != nil {
		if er := err.(*ProfileNotFoundError); er.Profile != NonExistingIdAsString {
			t.Fatalf("got unexpected error: %s", err)
		}
	}
}

func TestGetUserByIdIfExist_WithExistingId_MustReturnUserStruct(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	user, err := gapi.GetUserByIdIfExist(context.Background(), ExistingUserId)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if user == nil {
		t.Fatalf("got nil user")
	}
	if user.ID != ExistingUserId {
		t.Fatalf("got unexpected user ID %d", user.ID)
	}
}

func TestCheckIsRepoExistById_WithNonExistingRepoId_MustReturnFalse(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	isExisting, err := gapi.CheckIsRepoExistsById(context.Background(), NonExistingRepoId)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if isExisting {
		t.Fatalf("repo %d should not exist", NonExistingRepoId)
	}
}

func TestCheckIsRepoExistById_WithExistingRepoId_MustReturnTrue(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	isExisting, err := gapi.CheckIsRepoExistsById(context.Background(), ExistingRepoId)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !isExisting {
		t.Fatalf("repo %d should exist", ExistingRepoId)
	}
}

func TestGetRepoByIdIfExist_WithNonExistingRepoId_MustReturnNotFoundErr(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	_, err := gapi.GetRepoByIdIfExist(context.Background(), NonExistingRepoId)
	if err != nil {
		if er := err.(*RepoNotFoundError); er.Repo != NonExistingIdAsString {
			t.Fatalf("got unexpected error: %s", err)
		}
	}
}
func TestGetRepoByIdIfExist_WithExistingRepoId_MustReturnRepoStruct(t *testing.T) {
	gapi := NewDefaultGitHubAPI()
	repo, err := gapi.GetRepoByIdIfExist(context.Background(), ExistingRepoId)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if repo == nil {
		t.Fatalf("got nil repo")
	}

	if repo.ID != ExistingRepoId {
		t.Fatalf("got unexpected repo ID %d", repo.ID)
	}
}
