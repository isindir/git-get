/*
Copyright © 2020 Eriks Zelenka <isindir@users.sourceforge.net>

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
	"os"
	"path/filepath"

	"github.com/isindir/git-get/gitget"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var cfgFile string
var logLevel string
var stayOnRef bool
var shallow bool
var concurrencyLevel int

var levels = map[string]logrus.Level{
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
	Long: `'git-get' - all your project repositories

git-get clones/refreshes all you project repositories in
one go.

Yaml formatted configuration file specifies directory
structure of the project. git-get allows to create symlinks
to cloned repositories, clone one repository multiple time
having different directory name.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.Fatalln(err)
			os.Exit(1)
		}
		initLogging(logLevel)
		gitget.GitGetRepositories(cfgFile, concurrencyLevel, stayOnRef, shallow)
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
	rootCmd.Flags().StringVarP(&cfgFile, "config-file", "f", defaultValue, "configuration file")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Logging level [debug|info|warn|error|fatal|panic]")
	rootCmd.Flags().IntVarP(&concurrencyLevel, "concurrency-level", "c", 1, "Git get concurrnecy level")
	rootCmd.Flags().BoolVarP(&stayOnRef, "stay-on-ref", "t", false, "After refreshing repository from remote stay on ref branch")
	rootCmd.Flags().BoolVarP(&shallow, "shallow", "s", false, "Shallow clone, can be used in CI to fetch dependencies by ref")
}
