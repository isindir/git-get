//go:build !integration
// +build !integration

package gitget

import (
	"path"
	"testing"
)

var repoUrls = map[string]string{
	"https://github.com/ansible/ansible.git":               "ansible",
	"git@github.com:ansible/ansible.git":                   "ansible",
	"git@gitlab.com:devops/deploy/deployment-jobs.git":     "deployment-jobs",
	"https://gitlab.com/devops/deploy/deployment-jobs.git": "deployment-jobs",
}

type testGitGetRepo struct {
}

func TestSetDefaultRef(t *testing.T) {
	repo := Repo{}

	// Test setting to default if not specified
	repo.SetDefaultRef()
	if repo.Ref != "master" {
		t.Errorf("Expected 'master', got: '%s'", repo.Ref)
	}

	// Test that if already set - setting to default has no effect
	repo.Ref = "trunk"
	repo.SetDefaultRef()
	if repo.Ref != "trunk" {
		t.Errorf("Expected 'trunk', got: '%s'", repo.Ref)
	}
}

func TestGetRepoLocalName(t *testing.T) {

	// Test extracting altname from git repo name, if altname is not specified
	var altname string
	for repoURL, expectedAltName := range repoUrls {
		repo := Repo{}
		repo.URL = repoURL
		altname = repo.GetRepoLocalName()
		if altname != expectedAltName {
			t.Errorf("Expected '%s', got: '%s'", expectedAltName, altname)
		}
	}

	repo := Repo{}
	repo.AltName = "abc"
	altname = repo.GetRepoLocalName()
	if altname != "abc" {
		t.Errorf("Expected 'abc', got: '%s'", altname)
	}
}

func TestSetRepoLocalName(t *testing.T) {

	for repoURL, expectedAltName := range repoUrls {
		repo := Repo{}
		repo.URL = repoURL
		repo.SetRepoLocalName()
		if repo.AltName != expectedAltName {
			t.Errorf("Expected '%s', got: '%s'", expectedAltName, repo.AltName)
		}
	}

	repo := Repo{}
	repo.AltName = "abc"
	repo.SetRepoLocalName()
	if repo.AltName != "abc" {
		t.Errorf("Expected 'abc', got: '%s'", repo.AltName)
	}
}

func TestSetSha(t *testing.T) {
	repo := Repo{}
	repo.URL = "git@github.com:isindir/git-get.git"
	repo.Ref = "master"
	repo.fullPath = "/Users/erikszelenka/workspace/eriks/git-get/git-get"

	repo.SetSha()
	if repo.sha != "28f4e2d" {
		t.Errorf("Expected '28f4e2d', got: '%s'", repo.sha)
	}
}

func TestSetRepoFullPath(t *testing.T) {
	repo := Repo{
		Path:    "qqq",
		AltName: "abc",
	}
	expectedFullPath := path.Join("qqq", "abc")

	repo.SetRepoFullPath()
	if repo.fullPath != expectedFullPath {
		t.Errorf("Expected '%s', got: '%s'", expectedFullPath, repo.fullPath)
	}
}

func TestPathExists(t *testing.T) {
	// Test Happy path
	res, _ := PathExists(".")
	if !res {
		t.Errorf("Expected '.' path to exist, got: '%v'", res)
	}

	// Test path not found
	res, _ = PathExists("NotExpectedToFindMe")
	if res {
		t.Errorf("Expected 'NotExpectedToFindMe' path to NOT exist, got: '%v'", res)
	}
}

func TestRepoPathExists(t *testing.T) {
	// Test path exists
	repo := Repo{
		fullPath: ".",
	}

	if !repo.RepoPathExists() {
		t.Errorf("Expected '.' path to exist")
	}

	// Test path not found
	repo.fullPath = "NotExpectedToFindMe"
	if repo.RepoPathExists() {
		t.Errorf("Expected 'NotExpectedToFindMe' path to NOT exist")
	}
}

//(t *testing.T) {
