//go:build !integration
// +build !integration

package gitget

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"testing"

	execMocks "github.com/isindir/git-get/exec/mocks"
	gitgetMocks "github.com/isindir/git-get/gitget/mocks"
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
	repo.URL = "git@github.com:johndoe/git-get.git"
	repo.Ref = "master"
	repo.fullPath = "/Users/johndoe/workspace/john/git-get/git-get"

	repo.SetSha()
	assert.Equal(t, "8955354", repo.sha)
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

/*
func Test_Repo_IsRefTag(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		plainOpenError error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{Ref: "abc", fullPath: "cde_p"}, plainOpenError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{Ref: "cde", fullPath: "cde_a"}, plainOpenError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"show-ref", "--quiet", "--verify", fmt.Sprintf("refs/tags/%s", tc.repo.Ref)},
				(*bytes.Buffer)(nil),
				(*bytes.Buffer)(nil),
				tc.repo.fullPath).Return(exe, tc.plainOpenError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.IsRefTag()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
*/

func Test_Repo_IsRefBranch(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		plainOpenError error
		referenceError error
	}

	testCases := []testCase{
		{
			name:           "No errors",
			expectedResult: true,
			repo:           Repo{Ref: "abc", fullPath: "cde_p"},
			plainOpenError: nil,
			referenceError: nil,
		},
		{
			name:           "error on plan open",
			expectedResult: false,
			repo:           Repo{Ref: "cde", fullPath: "cde_a"},
			plainOpenError: fmt.Errorf("test error"),
			referenceError: nil,
		},
		{
			name:           "error on reference",
			expectedResult: false,
			repo:           Repo{Ref: "cde", fullPath: "cde_a"},
			plainOpenError: nil,
			referenceError: fmt.Errorf("test error"),
		},
		{
			name:           "both errors",
			expectedResult: false,
			repo:           Repo{Ref: "cde", fullPath: "cde_a"},
			plainOpenError: fmt.Errorf("test error"),
			referenceError: fmt.Errorf("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitgetRepo := new(gitgetMocks.RepoI)
			mockGoRepository := new(gitgetMocks.RepositoryI)
			mockGitgetRepo.On(
				"PlainOpen",
			).Return(mockGoRepository, tc.plainOpenError)

			mockGoRepository.On(
				"Reference",
			).Return(nil, tc.referenceError)

			result := tc.repo.IsRefBranch(mockGitgetRepo)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_GitStashPop(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		returnError    error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{}, returnError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"stash", "pop"},
				(*bytes.Buffer)(nil),
				serr,
				tc.repo.fullPath).Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.GitStashPop()
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, !tc.expectedResult, tc.repo.status.Error)
		})
	}
}

func Test_Repo_GitStashSave(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		returnError    error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{}, returnError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"stash", "save"},
				(*bytes.Buffer)(nil),
				serr,
				tc.repo.fullPath).Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.GitStashSave()
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, !tc.expectedResult, tc.repo.status.Error)
		})
	}
}

func Test_Repo_CloneMirror(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		returnError    error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{URL: "cde", fullPath: "cde_a"}, returnError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{URL: "cde", fullPath: "cde_a"}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"clone", "--mirror", tc.repo.URL, tc.repo.fullPath},
				(*bytes.Buffer)(nil),
				serr,
				"").Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.CloneMirror()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_GetCurrentBranch(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult string
		returnError    error
	}

	testCases := []testCase{
		// Have not found a way to mutate `outb` at mock call invocation time to validate against `expectedResult`
		{name: "test 1", expectedResult: "", repo: Repo{fullPath: "cde_a"}, returnError: nil},
		{name: "test 2", expectedResult: "", repo: Repo{fullPath: "cde_a"}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			var outb, errb bytes.Buffer

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"rev-parse", "--abbrev-ref", "HEAD"},
				&outb,
				&errb,
				tc.repo.fullPath).Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.GetCurrentBranch()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_IsClean(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		err1           error
		err2           error
	}

	testCases := []testCase{
		{name: "test 1: no error", expectedResult: true,
			repo: Repo{fullPath: "cde_p"}, err1: nil, err2: nil},
		{name: "test 2: err1 raised", expectedResult: false,
			repo: Repo{fullPath: "cde_a"}, err1: fmt.Errorf("test error"), err2: nil},
		{name: "test 3: err2 raised", expectedResult: false,
			repo: Repo{fullPath: "cde_a"}, err1: nil, err2: fmt.Errorf("test error")},
		{name: "test 3: err1 and err2 can be raised", expectedResult: false,
			repo: Repo{fullPath: "cde_a"}, err1: fmt.Errorf("test error"), err2: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"diff", "--quiet"},
				(*bytes.Buffer)(nil),
				(*bytes.Buffer)(nil),
				tc.repo.fullPath).Return(exe, tc.err1)
			mockGitExec.On(
				"ExecGitCommand",
				[]string{"diff", "--staged", "--quiet"},
				(*bytes.Buffer)(nil),
				(*bytes.Buffer)(nil),
				tc.repo.fullPath).Return(exe, tc.err2)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.IsClean()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_ShallowClone(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		returnError    error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{Ref: "abc", URL: "cde", fullPath: "cde_a"}, returnError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{Ref: "abc", URL: "cde", fullPath: "cde_a"}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"clone", "--depth", "1", "--branch", tc.repo.Ref, tc.repo.URL, tc.repo.fullPath},
				(*bytes.Buffer)(nil),
				serr,
				"").Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.ShallowClone()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_Clone(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		returnError    error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{Ref: "abc", URL: "cde", fullPath: "cde_a"}, returnError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{Ref: "abc", URL: "cde", fullPath: "cde_a"}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"clone", "--branch", tc.repo.Ref, tc.repo.URL, tc.repo.fullPath},
				(*bytes.Buffer)(nil),
				serr,
				"").Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.Clone()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_PushMirror(t *testing.T) {
	type testCase struct {
		name           string
		repo           Repo
		expectedResult bool
		returnError    error
	}

	testCases := []testCase{
		{name: "test 1", expectedResult: true, repo: Repo{mirrorURL: "cde"}, returnError: nil},
		{name: "test 2", expectedResult: false, repo: Repo{mirrorURL: "cde"}, returnError: fmt.Errorf("test error")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"push", "--mirror", tc.repo.mirrorURL},
				(*bytes.Buffer)(nil),
				serr,
				"").Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)

			result := tc.repo.PushMirror()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_Repo_GitCheckout(t *testing.T) {
	type testCase struct {
		name                   string
		repo                   Repo
		branch                 string
		expectedResult         bool
		expectedNotOnRefBranch bool
		returnError            error
	}

	testCases := []testCase{
		{
			name:                   "test 1",
			expectedResult:         true,
			expectedNotOnRefBranch: false,
			branch:                 "develop",
			repo:                   Repo{sha: "abcabc", Ref: "develop", fullPath: "abc"},
			returnError:            nil,
		},
		{
			name:                   "test 2",
			expectedResult:         false,
			expectedNotOnRefBranch: false,
			branch:                 "develop",
			repo:                   Repo{sha: "abcabc", Ref: "develop", fullPath: "abc"},
			returnError:            fmt.Errorf("test error"),
		},
		{
			name:                   "test 3",
			expectedResult:         true,
			expectedNotOnRefBranch: true,
			branch:                 "master",
			repo:                   Repo{sha: "abcabc", Ref: "develop", fullPath: "abc"},
			returnError:            nil,
		},
		{
			name:                   "test 4",
			expectedResult:         false,
			expectedNotOnRefBranch: true,
			branch:                 "master",
			repo:                   Repo{sha: "abcabc", Ref: "develop", fullPath: "abc"},
			returnError:            fmt.Errorf("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGitExec := new(execMocks.GitRunnerI)
			exe := &exec.Cmd{}
			serr := &bytes.Buffer{}

			mockGitExec.On(
				"ExecGitCommand",
				[]string{"checkout", tc.branch},
				(*bytes.Buffer)(nil),
				serr,
				tc.repo.fullPath).Return(exe, tc.returnError)

			tc.repo.SetGitRunner(mockGitExec)
			initColors()

			result := tc.repo.GitCheckout(tc.branch)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, !tc.expectedResult, tc.repo.status.Error)
			assert.Equal(t, tc.expectedNotOnRefBranch, tc.repo.status.NotOnRefBranch)
		})
	}
}
