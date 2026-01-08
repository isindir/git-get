/*
Copyright © 2021-2026 Eriks Zelenka <isindir@users.sourceforge.net>

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

package github

// UPDATE_HERE
import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v81/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type GitGetGithub struct {
	token string
}

type GitGetGithubI interface {
	Init() bool
	RepositoryExists(ctx context.Context, repositorySha, owner, repository string) bool
	CreateRepository(
		ctx context.Context,
		repositorySha string,
		repository string,
		mirrorVisibilityMode string,
		sourceURL string,
	) *github.Repository
	FetchOwnerRepos(
		ctx context.Context,
		repoSha, owner, githubVisibility, githubAffiliation string,
	) []*github.Repository
}

func (gitProvider *GitGetGithub) Init() bool {
	var tokenFound bool
	gitProvider.token, tokenFound = os.LookupEnv("GITHUB_TOKEN")
	if !tokenFound {
		log.Fatal("Error - environment variable GITHUB_TOKEN not found")
		os.Exit(1)
	}

	return tokenFound
}

func (gitProvider *GitGetGithub) auth(ctx context.Context, repositorySha string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitProvider.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	git := github.NewClient(tc)

	return git
}

// RepositoryExists - check if remote github repository exists (method)
func (gitProvider *GitGetGithub) RepositoryExists(ctx context.Context, repositorySha, owner, repository string) bool {
	git := gitProvider.auth(ctx, repositorySha)
	repo, _, err := git.Repositories.Get(ctx, owner, repository)

	log.Debugf("%s: %+v == %+v", repositorySha, repo, err)
	return err == nil
}

// RepositoryExists - check if remote github repository exists (package function for backward compatibility)
func RepositoryExists(ctx context.Context, repositorySha, owner, repository string) bool {
	gitProvider := &GitGetGithub{}
	gitProvider.Init()
	return gitProvider.RepositoryExists(ctx, repositorySha, owner, repository)
}

// CreateRepository - Create github repository (method)
func (gitProvider *GitGetGithub) CreateRepository(
	ctx context.Context,
	repositorySha string,
	repository string,
	mirrorVisibilityMode string,
	sourceURL string,
) *github.Repository {
	git := gitProvider.auth(ctx, repositorySha)
	isPrivate := true

	if mirrorVisibilityMode == "public" {
		isPrivate = false
	}

	repoDef := &github.Repository{
		Name:        github.Ptr(repository),
		Private:     github.Ptr(isPrivate),
		Description: github.Ptr(fmt.Sprintf("Mirror of the '%s'", sourceURL)),
	}

	resultingRepository, _, err := git.Repositories.Create(ctx, "", repoDef)
	if err != nil {
		log.Fatalf(
			"%s: Error - while trying to create github repository '%s': '%s'",
			repositorySha, repository, err)
		os.Exit(1)
	}

	return resultingRepository
}

// CreateRepository - Create github repository (package function for backward compatibility)
func CreateRepository(
	ctx context.Context,
	repositorySha string,
	repository string,
	mirrorVisibilityMode string,
	sourceURL string,
) *github.Repository {
	gitProvider := &GitGetGithub{}
	gitProvider.Init()
	return gitProvider.CreateRepository(ctx, repositorySha, repository, mirrorVisibilityMode, sourceURL)
}

func fetchOrgRepos(
	ctx context.Context,
	git *github.Client,
	repoSha, owner string,
) []*github.Repository {
	var repoList []*github.Repository

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page: 0,
		},
	}

	for {
		repos, res, err := git.Repositories.ListByOrg(ctx, owner, opts)
		log.Debugf(
			"%s: NextPage/PrevPage/FirstPage/LastPage '%d/%d/%d/%d'\n",
			repoSha, res.NextPage, res.PrevPage, res.FirstPage, res.LastPage)
		for repo := 0; repo < len(repos); repo++ {
			log.Debugf("%s: (%d) Repo FullName '%s'", repoSha, opts.ListOptions.Page, *repos[repo].FullName)
		}
		repoList = append(repoList, repos...)
		log.Debugf("%s: Found '%d' repositories owned by '%s'", repoSha, len(repos), owner)

		opts.ListOptions = github.ListOptions{
			Page: res.NextPage,
		}

		if res.NextPage == 0 {
			break
		}

		if err != nil {
			log.Debugf("%s: Error fetching repositories for '%s': %+v\n", repoSha, owner, err)
			break
		}
	}

	return repoList
}

func fetchUserRepos(
	ctx context.Context,
	git *github.Client,
	repoSha, owner, githubVisibility, githubAffiliation string,
) []*github.Repository {
	var repoList []*github.Repository

	opts := &github.RepositoryListByAuthenticatedUserOptions{
		// Default: all. Can be one of all, public, or private via CLI flags
		Visibility: githubVisibility,
		// Comma-separated list of values. Can include: owner, collaborator, or organization_member
		// Default: owner,collaborator,organization_member
		// Can be set via CLI flags
		Affiliation: githubAffiliation,
		ListOptions: github.ListOptions{
			Page: 0,
		},
	}

	for {
		repos, res, err := git.Repositories.ListByAuthenticatedUser(ctx, opts)
		log.Debugf(
			"%s: NextPage/PrevPage/FirstPage/LastPage '%d/%d/%d/%d'\n",
			repoSha, res.NextPage, res.PrevPage, res.FirstPage, res.LastPage)
		for repo := 0; repo < len(repos); repo++ {
			log.Debugf("%s: (%d) Repo FullName '%s'", repoSha, opts.ListOptions.Page, *repos[repo].FullName)
		}
		repoList = append(repoList, repos...)
		log.Debugf("%s: Found '%d' repositories owned by '%s'", repoSha, len(repos), owner)

		opts.ListOptions = github.ListOptions{
			Page: res.NextPage,
		}

		if res.NextPage == 0 {
			break
		}

		if err != nil {
			log.Debugf("%s: Error fetching repositories for '%s': %+v\n", repoSha, owner, err)
			break
		}
	}

	return repoList
}

// FetchOwnerRepos - fetch owner repositories via API, being it Organization or User (method)
func (gitProvider *GitGetGithub) FetchOwnerRepos(
	ctx context.Context,
	repoSha, owner, githubVisibility, githubAffiliation string,
) []*github.Repository {
	log.Debugf("%s: Specified owner: '%s'", repoSha, owner)
	git := gitProvider.auth(ctx, repoSha)
	var repoList []*github.Repository
	var userType string

	user, _, err := git.Users.Get(ctx, owner)
	if err != nil {
		log.Debugf("%s: Owner '%s' not found: '%+v'", repoSha, owner, err)
	} else {
		log.Debugf("%s: Owner '%s', Type: '%s'", repoSha, owner, *user.Type)
		userType = *user.Type
	}

	switch userType {
	case "Organization":
		repoList = fetchOrgRepos(ctx, git, repoSha, owner)
	case "User":
		repoList = fetchUserRepos(ctx, git, repoSha, owner, githubVisibility, githubAffiliation)
	default:
		log.Fatalf("%s: Error: unknown '%s' user type", repoSha, userType)
		os.Exit(1)
	}

	return repoList
}

// FetchOwnerRepos - fetch owner repositories via API, being it Organization or User (package function for backward compatibility)
func FetchOwnerRepos(
	ctx context.Context,
	repoSha, owner, githubVisibility, githubAffiliation string,
) []*github.Repository {
	gitProvider := &GitGetGithub{}
	gitProvider.Init()
	return gitProvider.FetchOwnerRepos(ctx, repoSha, owner, githubVisibility, githubAffiliation)
}
