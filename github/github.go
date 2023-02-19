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

package github

// UPDATE_HERE
import (
	"fmt"
	"os"

	"github.com/google/go-github/v50/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func githubAuth(ctx context.Context, repositorySha string) *github.Client {
	token, tokenFound := os.LookupEnv("GITHUB_TOKEN")
	if !tokenFound {
		log.Fatalf("%s: Error - environment variable GITHUB_TOKEN not found", repositorySha)
		os.Exit(1)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	git := github.NewClient(tc)

	return git
}

// RepositoryExists - check if remote github repository exists
func RepositoryExists(ctx context.Context, repositorySha string, owner string, repository string) bool {
	git := githubAuth(ctx, repositorySha)
	repo, _, err := git.Repositories.Get(ctx, owner, repository)

	log.Debugf("%s: %+v == %+v", repositorySha, repo, err)
	return err == nil
}

// CreateRepository - Create github repository
func CreateRepository(
	ctx context.Context,
	repositorySha string,
	repository string,
	mirrorVisibilityMode string,
	sourceURL string,
) *github.Repository {
	git := githubAuth(ctx, repositorySha)
	isPrivate := true

	if mirrorVisibilityMode == "public" {
		isPrivate = false
	}

	repoDef := &github.Repository{
		Name:        github.String(repository),
		Private:     github.Bool(isPrivate),
		Description: github.String(fmt.Sprintf("Mirror of the '%s'", sourceURL)),
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

	opts := &github.RepositoryListOptions{
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
		repos, res, err := git.Repositories.List(ctx, "", opts)
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

// FetchOwnerRepos - fetch owner repositories via API, being it Organization or User
func FetchOwnerRepos(
	ctx context.Context,
	repoSha, owner, githubVisibility, githubAffiliation string,
) []*github.Repository {
	log.Debugf("%s: Specified owner: '%s'", repoSha, owner)
	git := githubAuth(ctx, repoSha)
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
