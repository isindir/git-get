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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/isindir/git-get/gitget"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// configGenCmd represents the config-gen command
var configGenCmd = &cobra.Command{
	Use:   "config-gen",
	Short: "Create Gitfile configuration file from git provider",
	Long: `
Create 'Gitfile' configuration file dynamically from git provider by specifying
top level URL of the organisation organization, user or for Gitlab provider Group name.

* Github: Environment variable GITHUB_TOKEN defined.
* Bitbucket: Environment variables BITBUCKET_USERNAME and BITBUCKET_TOKEN (password) defined.
* Gitlab: Environment variable GITLAB_TOKEN defined.
* Gitlab: provider allows to create hierarchy of groups, 'git-get' is capable of fetching
  this hierarchy to 'Gifile' from any level visible to the user (see examples).`,
	Example: `
git-get config-gen -f Gitfile -p "gitlab" -u "git@gitlab.com:johndoe" -t misc -l debug
git-get config-gen -f Gitfile -p "gitlab" -u "git@gitlab.com:AcmeOrg" -t misc -l debug
git-get config-gen -f Gitfile -p "gitlab" -u "git@gitlab.com:AcmeOrg/kube"
git-get config-gen -f Gitfile -p "bitbucket" -u "git@bitbucket.com:AcmeOrg" -t AcmeOrg
git-get config-gen -f Gitfile -p "github" -u "git@github.com:johndoe" -t johndoe -l debug
git-get config-gen -f Gitfile -p "github" -u "git@github.com:AcmeOrg" -t AcmeOrg -l debug`,
	Run: func(cmd *cobra.Command, args []string) {
		initLogging(logLevel)
		log.Debug("Generate Gitfile configuration file")
		gitget.GenerateGitfileConfig(
			cfgFile, gitCloudProviderRootURL, gitCloudProvider, targetClonePath, configGenParams)
	},
}

func init() {
	rootCmd.AddCommand(configGenCmd)

	wdir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	defaultValue := filepath.Join(wdir, "Gitfile")
	defaultIgnoreValue := fmt.Sprintf("%s.ignore", defaultValue)
	configGenCmd.Flags().StringVarP(
		&cfgFile, "config-file",
		"f",
		defaultValue,
		"Configuration file")
	configGenCmd.Flags().StringVarP(
		&ignoreFile, "ignore-file",
		"i",
		defaultIgnoreValue,
		"Ignore file")
	configGenCmd.Flags().StringVarP(
		&logLevel,
		"log-level",
		"l",
		"info",
		"Logging level [debug|info|warn|error|fatal|panic]")
	configGenCmd.Flags().StringVarP(
		&gitCloudProvider,
		"config-provider",
		"p",
		"gitlab",
		"Git provider name [gitlab|github|bitbucket]")
	configGenCmd.Flags().StringVarP(
		&gitCloudProviderRootURL,
		"config-url",
		"u",
		"",
		"Private URL prefix to construct Gitfile from (example: git@github.com:acmeorg), provider specific.")
	configGenCmd.Flags().StringVarP(
		&targetClonePath,
		"target-clone-path",
		"t",
		"",
		"Target clone path used to set 'path' for each repository in Gitfile")
	configGenCmd.Flags().BoolVar(
		&configGenParams.GitlabOwned,
		"gitlab-owned",
		false,
		"Gitlab: only traverse groups and repositories owned by user")
	configGenCmd.Flags().StringVar(
		&configGenParams.GitlabVisibility,
		"gitlab-project-visibility",
		"",
		"Gitlab: project visibility [public|internal|private]")
	configGenCmd.Flags().StringVar(
		&configGenParams.GitlabMinAccessLevel,
		"gitlab-groups-minimal-access-level",
		"unspecified",
		"Gitlab: groups minimal access level [unspecified|min|guest|reporter|developer|maintainer|owner]")
	configGenCmd.Flags().StringVar(
		&configGenParams.GithubVisibility,
		"github-visibility",
		"all",
		"Github: visibility [all|public|private]")
	configGenCmd.Flags().StringVar(
		&configGenParams.GithubAffiliation,
		"github-affiliation",
		"owner,collaborator,organization_member",
		`Github: affiliation - comma-separated list of values.
Can include: owner, collaborator, or organization_member`)
	/*
		MAYBE: implement for bitbucket to allow subset of repositories
		configGenCmd.Flags().StringVar(
			&configGenParams.BitbucketDivision,
			"bitbucket-division",
			"account",
			"Bitbucket: Get repositories for [account|team]")
	*/
	configGenCmd.Flags().StringVar(
		&configGenParams.BitbucketRole,
		"bitbucket-role",
		"member",
		"Bitbucket: Filter repositories by role [owner|admin|contributor|member]")
}
