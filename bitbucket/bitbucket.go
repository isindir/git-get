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

package bitbucket

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	bitbucket "github.com/ktrysmt/go-bitbucket"
)

func bitbucketAuth(repoSha string) *bitbucket.Client {
	username, usernameFound := os.LookupEnv("BITBUCKET_USERNAME")
	if !usernameFound {
		log.Fatalf("%s: Error - environment variable BITBUCKET_TOKEN not found", repoSha)
		os.Exit(1)
	}

	token, tokenFound := os.LookupEnv("BITBUCKET_TOKEN")
	if !tokenFound {
		log.Fatalf("%s: Error - environment variable BITBUCKET_TOKEN not found", repoSha)
		os.Exit(1)
	}

	git := bitbucket.NewBasicAuth(username, token)

	return git
}

// GenerateProjectKey - convert project name to project key
func GenerateProjectKey(projectName string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return strings.ToUpper(re.ReplaceAllString(projectName, ""))
}

// RepositoryExists - checks if bitbucket repository exists
func RepositoryExists(repoSha string, owner string, repository string) bool {
	git := bitbucketAuth(repoSha)

	repoOptions := &bitbucket.RepositoryOptions{
		Owner:    owner,
		RepoSlug: repository,
	}
	repo, err := git.Repositories.Repository.Get(repoOptions)
	if err != nil {
		log.Debugf("%s: Error fetching repository '%s/%s': %+v", repoSha, owner, repository, err)
		return false
	}

	log.Debugf("%s: Fetched repository '%+v'", repoSha, repo)
	return true
}

// ProjectExists - checks if bitbucket project exists
func ProjectExists(git *bitbucket.Client, repoSha string, workspace string, project string) bool {
	opt := &bitbucket.ProjectOptions{
		Owner: workspace,
		Name:  project,
	}
	log.Debugf("%s: Parameter ProjectOptions '%+v'", repoSha, opt)

	prj, err := git.Workspaces.GetProject(opt)
	if err != nil {
		log.Debugf("%s: Error fetching project '%s' in workspace '%s': %+v\n", repoSha, project, workspace, err)
		return false
	}

	log.Debugf("%s: Fetched project '%+v'\n", repoSha, prj)
	return true
}

// CreateRepository - create bitbucket repository
func CreateRepository(repoSha, repository, mirrorVisibilityMode, sourceURL, projectName string) *bitbucket.Repository {
	git := bitbucketAuth(repoSha)

	repoNameParts := strings.SplitN(repository, "/", 2)
	owner, repoSlug := repoNameParts[0], repoNameParts[1]

	isPrivate := "true"
	if mirrorVisibilityMode == "public" {
		isPrivate = "false"
	}

	repoOptions := &bitbucket.RepositoryOptions{
		Owner:       owner,
		RepoSlug:    repoSlug,
		IsPrivate:   isPrivate,
		Description: fmt.Sprintf("Mirror of the '%s'", sourceURL),
		Scm:         "git",
	}
	// Assuming specific format of Key here - project name
	// is converted to uppercase and only [a-zA-Z0-9_] are allowed in input name
	if projectName != "" {
		projectKey := GenerateProjectKey(projectName)
		if ProjectExists(git, repoSha, owner, projectKey) {
			repoOptions.Project = projectKey
		}
	}
	log.Debugf("%s: Creating repository with parameters: '%+v'", repoSha, repoOptions)

	resultingRepository, err := git.Repositories.Repository.Create(repoOptions)
	if err != nil {
		log.Fatalf("%s: Error - while trying to create github repository '%s': '%s'", repoSha, repository, err)
		os.Exit(1)
	}

	log.Debugf("%s: Repository created: '%+v'", repoSha, resultingRepository)
	return resultingRepository
}

// FetchOwnerRepos - fetch owner repositories via API
func FetchOwnerRepos(repoSha, owner, bitbucketRole string) []bitbucket.Repository {
	log.Debugf("%s: Specified owner: '%s'", repoSha, owner)
	var reposToReutrn []bitbucket.Repository

	git := bitbucketAuth(repoSha)

	opts := &bitbucket.RepositoriesOptions{
		Owner: owner,
		Role:  bitbucketRole,
	}

	repos, err := git.Workspaces.Repositories.ListForAccount(opts)

	if repos != nil && len(repos.Items) > 0 && err == nil {
		log.Debugf(
			"%s: Page: %d, Pagelen: %d, Size: %d",
			repoSha, repos.Page, repos.Pagelen, repos.Size)
		reposToReutrn = repos.Items
	} else {
		log.Errorf("%s: Can't fetch repository list for '%s' '%+v'", repoSha, owner, err)
	}

	for i := 0; repos != nil && i < len(repos.Items) && err == nil; i++ {
		log.Debugf("%s: Repository '%s(%s)'", repoSha, repos.Items[i].Full_name, repos.Items[i].Mainbranch.Name)
	}

	return reposToReutrn
}
