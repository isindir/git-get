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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/isindir/git-get/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information",
	Run: func(cmd *cobra.Command, args []string) {

		long, _ := cmd.Flags().GetBool("long")

		if long == true {
			fmt.Printf(
				"%s %s %s %s %s/%s\n",
				filepath.Base(os.Args[0]),
				version.Version,
				version.Commit,
				runtime.Version(),
				runtime.GOOS,
				runtime.GOARCH,
			)
		} else {
			fmt.Printf(
				"%s %s\n",
				filepath.Base(os.Args[0]),
				version.Version,
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP(
		"long",
		"l",
		false,
		"print additional version information (default: false)",
	)
}
