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

func Test_SetDefaultRef(t *testing.T) {
	repo := Repo{}

	// Test setting to default if not specified
	repo.SetDefaultRef()
	assert.Equal(t, "master", repo.Ref)

	// Test that if already set - setting to default has no effect
	repo.Ref = "trunk"
	repo.SetDefaultRef()
	assert.Equal(t, "trunk", repo.Ref)
}

func Test_GetRepoLocalName(t *testing.T) {

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

func Test_SetRepoLocalName(t *testing.T) {

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

func Test_SetSha(t *testing.T) {
	repo := Repo{}
	repo.URL = "git@github.com:isindir/git-get.git"
	repo.Ref = "master"
	repo.fullPath = "/Users/erikszelenka/workspace/eriks/git-get/git-get"

	repo.SetSha()
	assert.Equal(t, "28f4e2d", repo.sha)
}

func Test_PathExists(t *testing.T) {
	// Test Happy path
	res, _ := PathExists(".")
	assert.True(t, res)

	// Test path not found
	res, _ = PathExists("NotExpectedToFindMe")
	assert.False(t, res)
}

func Test_SetRepoFullPath(t *testing.T) {
	repo := Repo{
		Path:    "qqq",
		AltName: "abc",
	}
	expectedFullPath := path.Join("qqq", "abc")

	repo.SetRepoFullPath()
	assert.Equal(t, expectedFullPath, repo.fullPath)
}

func Test_RepoPathExists(t *testing.T) {
	// Test path exists
	repo := Repo{
		fullPath: ".",
	}

	assert.True(t, repo.RepoPathExists())

	// Test path not found
	repo.fullPath = "NotExpectedToFindMe"
	assert.False(t, repo.RepoPathExists())
}

func Test_generateSha(t *testing.T) {
	type testCase struct {
		name           string
		repoInfo       string
		expectedResult string
	}

	testCases := []testCase{
		{name: "test 1", repoInfo: "Some string here", expectedResult: "21e5963"},
		{name: "test 2", repoInfo: "Another string here", expectedResult: "04d5c8f"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateSha(tc.repoInfo)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_SetMirrorURL(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult string
		mirrorRootURL  string
	}

	testCases := []testCase{
		{name: "test repo abc", expectedResult: "abc/abc.git", repo: Repo{AltName: "abc"}, mirrorRootURL: "abc"},
		{name: "test repo qqq", expectedResult: "abc/qqq/cde.git", repo: Repo{AltName: "cde"}, mirrorRootURL: "abc/qqq"},
		{name: "test repo http://qqq", expectedResult: "http://qqq/cde.git", repo: Repo{AltName: "cde"}, mirrorRootURL: "http://qqq"},
		{name: "test repo git@qqq", expectedResult: "git@qqq/cde.git", repo: Repo{AltName: "cde"}, mirrorRootURL: "git@qqq"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.repo.SetMirrorURL(tc.mirrorRootURL)
			assert.Equal(t, tc.expectedResult, tc.repo.mirrorURL)
		})
	}
}

//(t *testing.T) {
