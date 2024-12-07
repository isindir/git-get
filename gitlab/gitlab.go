/*
Copyright © 2021-2022 Eriks Zelenka <isindir@users.sourceforge.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package gitlab

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

type GitGetGitlab struct {
	token  string
	client *gitlab.Client
}

type GitGetGitlabI interface {
	Init() bool

	gitlabAuth(
		repositorySha string,
		baseUrl string,
	) bool
	ProjectExists(
		repositorySha string,
		baseUrl string,
		projectName string,
	) bool
	GetProjectNamespace(
		repositorySha string,
		baseUrl string,
		projectNameFullPath string,
	) (*gitlab.Namespace, string)
	getGroupID(
		repoSha string,
		git *gitlab.Client,
		groupName string,
	) (int, string, error)

	CreateProject(
		repositorySha string,
		baseUrl string,
		projectName string,
		namespaceID int,
		mirrorVisibilityMode string,
		sourceURL string,
	) *gitlab.Project

	processSubgroups(
		repoSha string,
		git *gitlab.Client,
		groupID int,
		groupName string,
		glRepoList []*gitlab.Project,
		gitlabOwned bool,
		gitlabVisibility, gitlabMinAccessLevel string,
	) []*gitlab.Project

	appendGroupsProjects(
		repoSha string,
		git *gitlab.Client,
		groupID int,
		groupName string,
		glRepoList []*gitlab.Project,
		gitlabOwned bool,
		gitlabVisibility string,
	) []*gitlab.Project

	FetchOwnerRepos(
		repositorySha, baseURL, groupName string,
		gitlabOwned bool,
		gitlabVisibility, gitlabMinAccessLevel string,
	) []*gitlab.Project
}

func (gitProvider *GitGetGitlab) Init() bool {
	var tokenFound bool
	gitProvider.token, tokenFound = os.LookupEnv("GITLAB_TOKEN")
	if !tokenFound {
		log.Fatal("Error - environment variable GITLAB_TOKEN not found")
		os.Exit(1)
	}

	return tokenFound
}

func (gitProvider *GitGetGitlab) auth(repositorySha string, baseUrl string) bool {
	clientOptions := gitlab.WithBaseURL("https://" + baseUrl)
	var err error
	gitProvider.client, err = gitlab.NewClient(gitProvider.token, clientOptions)
	if err != nil {
		log.Fatalf("%s: Error - while trying to authenticate to Gitlab: %s", repositorySha, err)
		os.Exit(1)
	}

	return true
}

// ProjectExists checks if project exists and returns boolean if API call is successful
func (gitProvider *GitGetGitlab) ProjectExists(repositorySha string, baseUrl string, projectName string) bool {
	log.Debugf("%s: Checking repository '%s' '%s' existence", repositorySha, baseUrl, projectName)
	gitProvider.auth(repositorySha, baseUrl)

	prj, _, err := gitProvider.client.Projects.GetProject(projectName, nil, nil)

	log.Debugf("%s: project: '%+v'", repositorySha, prj)

	return err == nil
}

// GetProjectNamespace - return Project Namespace and namespace full path
func (gitProvider *GitGetGitlab) GetProjectNamespace(
	repositorySha string,
	baseUrl string,
	projectNameFullPath string,
) (*gitlab.Namespace, string) {
	log.Debugf("%s: Getting Project FullPath Namespace '%s'", repositorySha, projectNameFullPath)
	gitProvider.auth(repositorySha, baseUrl)

	pathElements := strings.Split(projectNameFullPath, "/")
	// Remove short project name from the path elements list
	if len(pathElements) > 0 {
		pathElements = pathElements[:len(pathElements)-1]
	}
	namespaceFullPath := strings.Join(pathElements, "/")

	namespaceObject, _, err := gitProvider.client.Namespaces.GetNamespace(namespaceFullPath, nil, nil)

	log.Debugf(
		"%s: Getting namespace '%s': resulting namespace object: '%+v'",
		repositorySha, namespaceFullPath, namespaceObject)

	if err != nil {
		return nil, namespaceFullPath
	}

	return namespaceObject, namespaceFullPath
}

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func gitlabVisibilityValuePtr(value gitlab.VisibilityValue) *gitlab.VisibilityValue {
	return &value
}

// CreateProject - Create new code repository
func (gitProvider *GitGetGitlab) CreateProject(
	repositorySha string,
	baseUrl string,
	projectName string,
	namespaceID int,
	mirrorVisibilityMode string,
	sourceURL string,
) *gitlab.Project {
	gitProvider.auth(repositorySha, baseUrl)

	p := &gitlab.CreateProjectOptions{
		Name: stringPtr(projectName),
		Description: stringPtr(
			fmt.Sprintf("Mirror of the '%s'", sourceURL),
		),
		MergeRequestsEnabled: boolPtr(true),
		Visibility: gitlabVisibilityValuePtr(
			gitlab.VisibilityValue(mirrorVisibilityMode),
		),
	}
	if namespaceID != 0 {
		p.NamespaceID = intPtr(namespaceID)
	}

	project, _, err := gitProvider.client.Projects.CreateProject(p)
	if err != nil {
		log.Fatalf(
			"%s: Error - while trying to create gitlab project '%s': '%s'",
			repositorySha,
			projectName,
			err,
		)
		os.Exit(1)
	}

	return project
}

func (gitProvider *GitGetGitlab) getGroupID(
	repoSha string,
	git *gitlab.Client,
	groupName string,
) (int, string, error) {
	// Fetch group ID needed for other operations
	_, shortName := filepath.Split(groupName)
	// escapedGroupName := url.QueryEscape(groupName)
	escapedGroupName := url.QueryEscape(shortName)

	foundGroups, _, err := git.Groups.SearchGroup(escapedGroupName)
	if err != nil {
		return 0, "", err
	}
	log.Debugf("%s: Found '%d' groups for '%s'", repoSha, len(foundGroups), groupName)
	for group := 0; group < len(foundGroups); group++ {
		log.Debugf(
			"%s: Checking group '%d:%s' to be as specified",
			repoSha,
			foundGroups[group].ID,
			foundGroups[group].FullName,
		)
		if groupName == strings.Replace(foundGroups[group].FullName, " ", "", -1) {
			return foundGroups[group].ID, foundGroups[group].FullName, nil
		}
	}

	return 0, "", fmt.Errorf("%s: Group with name '%s' not found", repoSha, groupName)
}

func (gitProvider *GitGetGitlab) processSubgroups(
	repoSha string,
	git *gitlab.Client,
	groupID int,
	groupName string,
	glRepoList []*gitlab.Project,
	gitlabOwned bool,
	gitlabVisibility, gitlabMinAccessLevel string,
) []*gitlab.Project {
	// Fetch all subgroups of this group and recurse
	var subGroups []*gitlab.Group

	lstOpts := gitlab.ListOptions{
		Page:    0,
		PerPage: 30,
	}

	subGrpOpt := &gitlab.ListSubGroupsOptions{
		ListOptions:  lstOpts,
		AllAvailable: boolPtr(true),
		TopLevelOnly: boolPtr(false),
		Owned:        boolPtr(gitlabOwned),
	}
	switch gitlabMinAccessLevel {
	case "min":
		minAccessLevel := gitlab.MinimalAccessPermissions
		subGrpOpt.MinAccessLevel = &minAccessLevel
	case "guest":
		minAccessLevel := gitlab.GuestPermissions
		subGrpOpt.MinAccessLevel = &minAccessLevel
	case "reporter":
		minAccessLevel := gitlab.ReporterPermissions
		subGrpOpt.MinAccessLevel = &minAccessLevel
	case "developer":
		minAccessLevel := gitlab.DeveloperPermissions
		subGrpOpt.MinAccessLevel = &minAccessLevel
	case "maintainer":
		minAccessLevel := gitlab.MaintainerPermissions
		subGrpOpt.MinAccessLevel = &minAccessLevel
	case "owner":
		minAccessLevel := gitlab.OwnerPermissions
		subGrpOpt.MinAccessLevel = &minAccessLevel
	default:
		log.Debugf("%s: Won't set MinAccessLevel - value passed is '%s'", repoSha, gitlabMinAccessLevel)
	}

	for {
		log.Debugf(
			"%s: Fetching group '%d:%s' subgroups - page '%d'",
			repoSha, groupID, groupName, subGrpOpt.ListOptions.Page)
		groups, res, err := git.Groups.ListSubGroups(groupID, subGrpOpt, nil)
		if err != nil {
			log.Errorf(
				"%s: Error while trying to get subgroups for '%d:%s': %s",
				repoSha, groupID, groupName, err,
			)
			break
		}
		log.Debugf(
			"%s: NextPage/PrevPage/CurrentPage/TotalPages '%d/%d/%d/%d'\n",
			repoSha, res.NextPage, res.PreviousPage, res.CurrentPage, res.TotalPages)

		subGroups = append(subGroups, groups...)
		log.Debugf("%s: Page %d: subgroups: %+v", repoSha, res.CurrentPage, subGroups)
		log.Debugf("%s: Page %d: current subgroups %+v", repoSha, res.CurrentPage, groups)

		// reached the end of page list
		if res.NextPage == 0 {
			break
		}

		// Prepare pagination options for next request
		lstOpts.Page = res.NextPage
		subGrpOpt.ListOptions = lstOpts
	}

	log.Debugf("%s: %+v", repoSha, subGroups)
	for currentGroup := 0; currentGroup < len(subGroups); currentGroup++ {
		log.Debugf(
			"%s: Recursive getRepositories call for subgroup '%d:%s'",
			repoSha, subGroups[currentGroup].ID, subGroups[currentGroup].FullName)

		glRepoList = gitProvider.getRepositories(
			repoSha,
			git,
			subGroups[currentGroup].ID,
			subGroups[currentGroup].FullName,
			glRepoList,
			gitlabOwned,
			gitlabVisibility,
			gitlabMinAccessLevel,
		)
	}

	return glRepoList
}

func (gitProvider *GitGetGitlab) appendGroupsProjects(
	repoSha string,
	git *gitlab.Client,
	groupID int,
	groupName string,
	glRepoList []*gitlab.Project,
	gitlabOwned bool,
	gitlabVisibility string,
) []*gitlab.Project {
	// Fetch all subgroup projects
	var subGroupProjects []*gitlab.Project

	lstOpts := gitlab.ListOptions{
		Page:    0,
		PerPage: 30,
	}

	// Fetch all projects(repositories) of the group and append to main list
	// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects
	prjOpt := &gitlab.ListGroupProjectsOptions{
		ListOptions: lstOpts,
		Owned:       boolPtr(gitlabOwned),
		Simple:      boolPtr(true),
	}
	switch gitlabVisibility {
	case "private":
		vis := gitlab.PrivateVisibility
		prjOpt.Visibility = &vis
	case "internal":
		vis := gitlab.InternalVisibility
		prjOpt.Visibility = &vis
	case "public":
		vis := gitlab.PublicVisibility
		prjOpt.Visibility = &vis
	default:
		log.Debugf("%s: Won't set Visibility - value passed is '%s'", repoSha, gitlabVisibility)
	}

	for {
		projects, res, err := git.Groups.ListGroupProjects(groupID, prjOpt)
		if err != nil {
			log.Errorf(
				"%s: Error while fetching groups '%s' repositories: %s",
				repoSha, groupName, err,
			)
			break
		}
		log.Debugf(
			"%s: NextPage/PrevPage/CurrentPage/TotalPages '%d/%d/%d/%d'\n",
			repoSha, res.NextPage, res.PreviousPage, res.CurrentPage, res.TotalPages)

		subGroupProjects = append(subGroupProjects, projects...)

		log.Debugf("%s: Page %d: subprojects: %+v", repoSha, res.CurrentPage, subGroupProjects)
		log.Debugf("%s: Page %d: current projects %+v", repoSha, res.CurrentPage, projects)

		// reached the end of page list
		if res.NextPage == 0 {
			break
		}

		// Prepare pagination options for next request
		lstOpts.Page = res.NextPage
		prjOpt.ListOptions = lstOpts
	}

	glRepoList = append(glRepoList, subGroupProjects...)
	log.Debugf("%s: Collected projects len: '%d'", repoSha, len(glRepoList))

	return glRepoList
}

// Recursive function via processSubgroups
func (gitProvider *GitGetGitlab) getRepositories(
	repoSha string,
	git *gitlab.Client,
	groupID int,
	groupName string,
	glRepoList []*gitlab.Project,
	gitlabOwned bool,
	gitlabVisibility, gitlabMinAccessLevel string,
) []*gitlab.Project {
	log.Debugf("%s: Ready to start processing subgroups for '%d:%s'", repoSha, groupID, groupName)
	glRepoList = gitProvider.processSubgroups(
		repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility, gitlabMinAccessLevel)

	log.Debugf("%s: Ready to start processing projects for '%d:%s'", repoSha, groupID, groupName)
	glRepoList = gitProvider.appendGroupsProjects(
		repoSha, git, groupID, groupName, glRepoList, gitlabOwned, gitlabVisibility)

	return glRepoList
}

// FetchOwnerRepos - fetches all repositories for the specified gitlab path
func (gitProvider *GitGetGitlab) FetchOwnerRepos(
	repositorySha, baseURL, groupName string,
	gitlabOwned bool,
	gitlabVisibility, gitlabMinAccessLevel string,
) []*gitlab.Project {
	gitProvider.auth(repositorySha, baseURL)
	var glRepoList []*gitlab.Project

	log.Debugf("%s: Get groupID for '%s'", repositorySha, groupName)
	groupID, fullGroupName, err := gitProvider.getGroupID(repositorySha, gitProvider.client, groupName)
	if err != nil {
		log.Errorf(
			"%s: Error while trying to find group '%s': %s",
			repositorySha, groupName, err,
		)
		return glRepoList
	}
	log.Debugf("%s: GroupID for '%s' is '%d'", repositorySha, fullGroupName, groupID)

	glRepoList = gitProvider.getRepositories(
		repositorySha,
		gitProvider.client,
		groupID,
		fullGroupName,
		glRepoList,
		gitlabOwned,
		gitlabVisibility,
		gitlabMinAccessLevel,
	)

	return glRepoList
}
