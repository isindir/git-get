// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	gitlab "github.com/xanzy/go-gitlab"
)

// GitGetGitlabI is an autogenerated mock type for the GitGetGitlabI type
type GitGetGitlabI struct {
	mock.Mock
}

// CreateProject provides a mock function with given fields: repositorySha, baseUrl, projectName, namespaceID, mirrorVisibilityMode, sourceURL
func (_m *GitGetGitlabI) CreateProject(repositorySha string, baseUrl string, projectName string, namespaceID int, mirrorVisibilityMode string, sourceURL string) *gitlab.Project {
	ret := _m.Called(repositorySha, baseUrl, projectName, namespaceID, mirrorVisibilityMode, sourceURL)

	if len(ret) == 0 {
		panic("no return value specified for CreateProject")
	}

	var r0 *gitlab.Project
	if rf, ok := ret.Get(0).(func(string, string, string, int, string, string) *gitlab.Project); ok {
		r0 = rf(repositorySha, baseUrl, projectName, namespaceID, mirrorVisibilityMode, sourceURL)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Project)
		}
	}

	return r0
}

// FetchOwnerRepos provides a mock function with given fields: repositorySha, baseURL, groupName, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel
func (_m *GitGetGitlabI) FetchOwnerRepos(repositorySha string, baseURL string, groupName string, gitlabOwned bool, gitlabVisibility string, gitlabMinAccessLevel string) []*gitlab.Project {
	ret := _m.Called(repositorySha, baseURL, groupName, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel)

	if len(ret) == 0 {
		panic("no return value specified for FetchOwnerRepos")
	}

	var r0 []*gitlab.Project
	if rf, ok := ret.Get(0).(func(string, string, string, bool, string, string) []*gitlab.Project); ok {
		r0 = rf(repositorySha, baseURL, groupName, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.Project)
		}
	}

	return r0
}

// GetProjectNamespace provides a mock function with given fields: repositorySha, baseUrl, projectNameFullPath
func (_m *GitGetGitlabI) GetProjectNamespace(repositorySha string, baseUrl string, projectNameFullPath string) (*gitlab.Namespace, string) {
	ret := _m.Called(repositorySha, baseUrl, projectNameFullPath)

	if len(ret) == 0 {
		panic("no return value specified for GetProjectNamespace")
	}

	var r0 *gitlab.Namespace
	var r1 string
	if rf, ok := ret.Get(0).(func(string, string, string) (*gitlab.Namespace, string)); ok {
		return rf(repositorySha, baseUrl, projectNameFullPath)
	}
	if rf, ok := ret.Get(0).(func(string, string, string) *gitlab.Namespace); ok {
		r0 = rf(repositorySha, baseUrl, projectNameFullPath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gitlab.Namespace)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string) string); ok {
		r1 = rf(repositorySha, baseUrl, projectNameFullPath)
	} else {
		r1 = ret.Get(1).(string)
	}

	return r0, r1
}

// Init provides a mock function with no fields
func (_m *GitGetGitlabI) Init() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Init")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ProjectExists provides a mock function with given fields: repositorySha, baseUrl, projectName
func (_m *GitGetGitlabI) ProjectExists(repositorySha string, baseUrl string, projectName string) bool {
	ret := _m.Called(repositorySha, baseUrl, projectName)

	if len(ret) == 0 {
		panic("no return value specified for ProjectExists")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, string) bool); ok {
		r0 = rf(repositorySha, baseUrl, projectName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// appendGroupsProjects provides a mock function with given fields: repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility
func (_m *GitGetGitlabI) appendGroupsProjects(repoSha string, git *gitlab.Client, groupID int, groupName string, glRepoList []*gitlab.Project, gitlabOwned bool, gitlabVisibility string) []*gitlab.Project {
	ret := _m.Called(repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility)

	if len(ret) == 0 {
		panic("no return value specified for appendGroupsProjects")
	}

	var r0 []*gitlab.Project
	if rf, ok := ret.Get(0).(func(string, *gitlab.Client, int, string, []*gitlab.Project, bool, string) []*gitlab.Project); ok {
		r0 = rf(repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.Project)
		}
	}

	return r0
}

// getGroupID provides a mock function with given fields: repoSha, git, groupName
func (_m *GitGetGitlabI) getGroupID(repoSha string, git *gitlab.Client, groupName string) (int, string, error) {
	ret := _m.Called(repoSha, git, groupName)

	if len(ret) == 0 {
		panic("no return value specified for getGroupID")
	}

	var r0 int
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(string, *gitlab.Client, string) (int, string, error)); ok {
		return rf(repoSha, git, groupName)
	}
	if rf, ok := ret.Get(0).(func(string, *gitlab.Client, string) int); ok {
		r0 = rf(repoSha, git, groupName)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(string, *gitlab.Client, string) string); ok {
		r1 = rf(repoSha, git, groupName)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string, *gitlab.Client, string) error); ok {
		r2 = rf(repoSha, git, groupName)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// gitlabAuth provides a mock function with given fields: repositorySha, baseUrl
func (_m *GitGetGitlabI) gitlabAuth(repositorySha string, baseUrl string) bool {
	ret := _m.Called(repositorySha, baseUrl)

	if len(ret) == 0 {
		panic("no return value specified for gitlabAuth")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(repositorySha, baseUrl)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// processSubgroups provides a mock function with given fields: repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel
func (_m *GitGetGitlabI) processSubgroups(repoSha string, git *gitlab.Client, groupID int, groupName string, glRepoList []*gitlab.Project, gitlabOwned bool, gitlabVisibility string, gitlabMinAccessLevel string) []*gitlab.Project {
	ret := _m.Called(repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel)

	if len(ret) == 0 {
		panic("no return value specified for processSubgroups")
	}

	var r0 []*gitlab.Project
	if rf, ok := ret.Get(0).(func(string, *gitlab.Client, int, string, []*gitlab.Project, bool, string, string) []*gitlab.Project); ok {
		r0 = rf(repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*gitlab.Project)
		}
	}

	return r0
}

// NewGitGetGitlabI creates a new instance of GitGetGitlabI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGitGetGitlabI(t interface {
	mock.TestingT
	Cleanup(func())
}) *GitGetGitlabI {
	mock := &GitGetGitlabI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
