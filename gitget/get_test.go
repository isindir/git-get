//go:build !integration
// +build !integration

package gitget

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var repoUrls = map[string]string{
	"https://github.com/ansible/ansible.git":               "ansible",
	"git@github.com:ansible/ansible.git":                   "ansible",
	"git@gitlab.com:devops/deploy/deployment-jobs.git":     "deployment-jobs",
	"https://gitlab.com/devops/deploy/deployment-jobs.git": "deployment-jobs",
}

func TestSetDefaultRef(t *testing.T) {
	repo := Repo{}

	// Test setting to default if not specified
	repo.SetDefaultRef()
	assert.Equal(t, "master", repo.Ref)

	// Test that if already set - setting to default has no effect
	repo.Ref = "trunk"
	repo.SetDefaultRef()
	assert.Equal(t, "trunk", repo.Ref)
}

func TestGetRepoLocalName(t *testing.T) {

	// Test extracting altname from git repo name, if altname is not specified
	var altname string
	for repoURL, expectedAltName := range repoUrls {
		repo := Repo{}
		repo.URL = repoURL
		altname = repo.GetRepoLocalName()
		assert.Equal(t, expectedAltName, altname)
	}

	repo := Repo{}
	repo.AltName = "abc"
	altname = repo.GetRepoLocalName()
	assert.Equal(t, "abc", altname)
}

func TestSetRepoLocalName(t *testing.T) {

	for repoURL, expectedAltName := range repoUrls {
		repo := Repo{}
		repo.URL = repoURL
		repo.SetRepoLocalName()
		assert.Equal(t, expectedAltName, repo.AltName)
	}

	repo := Repo{}
	repo.AltName = "abc"
	repo.SetRepoLocalName()
	assert.Equal(t, "abc", repo.AltName)
}

func TestSetSha(t *testing.T) {
	repo := Repo{}
	repo.URL = "git@github.com:isindir/git-get.git"
	repo.Ref = "master"
	repo.fullPath = "/Users/erikszelenka/workspace/eriks/git-get/git-get"

	repo.SetSha()
	assert.Equal(t, "28f4e2d", repo.sha)
}

func TestSetRepoFullPath(t *testing.T) {
	repo := Repo{
		Path:    "qqq",
		AltName: "abc",
	}
	expectedFullPath := path.Join("qqq", "abc")

	repo.SetRepoFullPath()
	assert.Equal(t, expectedFullPath, repo.fullPath)
}

func TestPathExists(t *testing.T) {
	// Test Happy path
	res, _ := PathExists(".")
	assert.True(t, res)

	// Test path not found
	res, _ = PathExists("NotExpectedToFindMe")
	assert.False(t, res)
}

func TestRepoPathExists(t *testing.T) {
	// Test path exists
	repo := Repo{
		fullPath: ".",
	}

	assert.True(t, repo.RepoPathExists())

	// Test path not found
	repo.fullPath = "NotExpectedToFindMe"
	assert.False(t, repo.RepoPathExists())
}

//(t *testing.T) {
