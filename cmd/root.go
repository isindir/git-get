/*
Copyright © 2020-2022 Eriks Zelenka <isindir@users.sourceforge.net>

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
	"fmt"
	"os"
	"path/filepath"

	"github.com/isindir/git-get/gitget"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var configGenParams gitget.ConfigGenParamsStruct

var cfgFile string
var cfgFiles []string
var ignoreFiles []string
var logLevel string
var stayOnRef bool
var shallow bool
var concurrencyLevel int
var pushMirror bool
var dryRun bool
var gitCloudProviderRootURL string
var targetClonePath string
var defaultMainBranch string
var status bool
var gitCloudProvider string

// Mirroring specific vars
var mirrorVisibilityMode string
var mirrorBitbucketProjectName string

var levels = map[string]log.Level{
	"panic": log.PanicLevel,
	"fatal": log.FatalLevel,
	"error": log.ErrorLevel,
	"warn":  log.WarnLevel,
	"info":  log.InfoLevel,
	"debug": log.DebugLevel,
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-get",
	Short: "'git-get' - all your project repositories",
	Long: `
'git-get' - all your project repositories

git-get clone/refresh all your local project repositories in
one go.

Yaml formatted configuration file specifies directory
structure of the project. git-get allows to create symlinks
to cloned repositories, clone one repository multiple time
having different directory name.`,
	Example: `
git get -c 12 -f Gitfile
git get -c 12 -f Gitfile.1 -f Gitfile.2 -f Gitfile.3,Gitfile.4
git get -c 12 -f Gitfile -i Gitfile.ignore.1 -i Gitfile.ignore.2
git get -c 8 -f Gitfile --status -i Gitfile.ignore -l panic \
  | awk '$0 ~ /REPOSITORY/ || $3 ~ /true/ { print $0 }'`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, cfgFile := range cfgFiles {
			if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
				log.Fatalln(err)
				os.Exit(1)
			}
		}
		initLogging(logLevel)
		gitget.GetRepositories(
			cfgFiles,
			ignoreFiles,
			concurrencyLevel,
			stayOnRef,
			shallow,
			defaultMainBranch,
			status,
		)
	},
}

func initLogging(level string) {
	log.SetLevel(levels[logLevel])
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func init() {
	wdir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	defaultValue := filepath.Join(wdir, "Gitfile")
	defaultIgnoreValue := fmt.Sprintf("%s.ignore", defaultValue)
	rootCmd.Flags().StringSliceVarP(
		&cfgFiles, "config-file",
		"f",
		[]string{defaultValue},
		"Configuration file or comma separated list of files")
	rootCmd.Flags().StringSliceVarP(
		&ignoreFiles, "ignore-file",
		"i",
		[]string{defaultIgnoreValue},
		"Ignore file or comma separated list of files")
	rootCmd.Flags().StringVarP(
		&logLevel, "log-level",
		"l",
		"info",
		"Logging level [debug|info|warn|error|fatal|panic]")
	rootCmd.Flags().IntVarP(
		&concurrencyLevel, "concurrency-level",
		"c",
		1,
		"Git get concurrency level")
	rootCmd.Flags().BoolVarP(
		&stayOnRef, "stay-on-ref",
		"t",
		false,
		"After refreshing repository from remote stay on ref branch")
	rootCmd.Flags().BoolVarP(
		&shallow, "shallow",
		"s",
		false,
		"Shallow clone, can be used in CI to fetch dependencies by ref")
	rootCmd.Flags().BoolVar(
		&status, "status",
		false,
		"Print extra status information after clone is performed")
	rootCmd.Flags().StringVarP(
		&defaultMainBranch, "default-main-branch",
		"b",
		"master",
		"Default main branch")
}
