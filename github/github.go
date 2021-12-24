/*
Copyright © 2021 Eriks Zelenka <isindir@users.sourceforge.net>

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
	"github.com/google/go-github/v41/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"os"
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

func RepositoryExists(ctx context.Context, repositorySha string, owner string, repository string) bool {
	git := githubAuth(ctx, repositorySha)
	repo, _, err := git.Repositories.Get(ctx, owner, repository)

	log.Debugf("%s: %+v == %+v", repositorySha, repo, err)
	return err == nil
}

func CreateRepository(ctx context.Context, repositorySha string, repository string, mirrorVisibilityMode string, sourceURL string) *github.Repository {
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
		log.Fatalf("%s: Error - while trying to create github repository '%s': '%s'", repositorySha, repository, err)
		os.Exit(1)
	}

	return resultingRepository
}
