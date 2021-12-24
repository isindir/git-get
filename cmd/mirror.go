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

// Package cmd provides logic for cli entrance point.
package cmd

import (
	"github.com/isindir/git-get/gitget"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// mirrorCmd represents the mirror command
var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Creates or updates a mirror of repositories",
	Long: `
Creates or updates a mirror of repositories specified by configuration file in a specified git provider cloud.

Different git providers have different workspace/project/organization/team/user/repository
structure/terminlogy/relations.

Notes:

* All providers: ssh key is used to clone/push git repositories, where environment
  variables are used to interrogate API.
* Gitlab: ssh key configured and environment variable GITLAB_TOKEN defined.
* Github: ssh key configured and environment variable GITHUB_TOKEN defined.
* Bitbucket: ssh key configured and environment variables BITBUCKET_USERNAME and BITBUCKET_TOKEN (password) defined.
* Bitbucket: Application won't create Project in Bitbucket if project is specified but missing.
  It assumes the Key of project to be constructed from it's name as Uppercase text containing
  only [A-Z0-9_] characters, all the rest of the characters from Project Name will be removed.`,
	Example: `
git get mirror -f Gitfile -u "git@github.com:acmeorg" -m "github"
git-get mirror -c 2 -f Gitfile -l debug -u "git@gitlab.com:acmeorg/mirrors"
git-get mirror -c 2 -f Gitfile -l debug -u "git@bitbucket.com:acmeorg" -m "bitbucket" -b "mirrors"`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.Fatalln(err)
			os.Exit(1)
		}
		initLogging(logLevel)
		log.Debugf("%t - push to mirror", pushMirror)
		gitget.MirrorRepositories(cfgFile, concurrencyLevel, pushMirror, mirrorRootURL, mirrorProvider, mirrorVisibilityMode, mirrorBitbucketProjectName)
	},
}

func init() {
	rootCmd.AddCommand(mirrorCmd)

	wdir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	defaultValue := filepath.Join(wdir, "Gitfile")
	mirrorCmd.Flags().StringVarP(&cfgFile, "config-file", "f", defaultValue, "Configuration file")
	mirrorCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Logging level [debug|info|warn|error|fatal|panic]")
	mirrorCmd.Flags().IntVarP(&concurrencyLevel, "concurrency-level", "c", 1, "Git get concurrency level")
	mirrorCmd.Flags().BoolVarP(&pushMirror, "push", "p", true, "Push to remote mirror repositories")
	mirrorCmd.Flags().StringVarP(&mirrorRootURL, "mirror-url", "u", "", "Private Mirror URL prefix to push repositories to (example: git@github.com:acmeorg)")
	mirrorCmd.Flags().StringVarP(&mirrorProvider, "mirror-provider", "m", "gitlab", "Git mirror provider name [gitlab|github|bitbucket]")
	mirrorCmd.Flags().StringVarP(&mirrorVisibilityMode, "mirror-visibility-mode", "v", "private", "Mirror visibility mode [private|internal|public]")
	mirrorCmd.Flags().StringVarP(&mirrorBitbucketProjectName, "bitbucket-mirror-project-name", "b", "", "Bitbucket mirror project name (only effective for Bitbucket and is optional)")
}
