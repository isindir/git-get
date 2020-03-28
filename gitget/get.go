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
package gitget

//go:generate moq -out gitgetrepoi_moq_test.go . GitGetRepoI

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const gitCmd = "git"

var stayOnRef bool
var colorHighlight *color.Color
var colorRef *color.Color

type GitGetRepo struct {
	Url      string   `yaml:"url"`
	Path     string   `yaml:"path,omitempty"`
	AltName  string   `yaml:"altname,omitempty"`
	Ref      string   `yaml:"ref,omitempty"`
	Symlinks []string `yaml:"symlinks,omitempty"`
	fullPath string
	sha      string
}

type GitGetRepoI interface {
	Clone() bool
	CreateSymlink(symlink string)
	EnsurePathExists()
	ExecGitCommand(args []string, stdoutb *bytes.Buffer, erroutb *bytes.Buffer, dir string) (cmd *exec.Cmd, err error)
	GetCurrentBranch() string
	GetRepoLocalName() string
	GitCheckout(branch string)
	GitPull()
	GitStashPop()
	GitStashSave()
	IsClean() bool
	IsCurrentBranchRef() bool
	IsRefBranch() bool
	IsRefTag() bool
	PathExists(path string) (bool, os.FileInfo)
	Prepare()
	ProcessRepoBasedOnCleaness()
	ProcessRepoBasedOnCurrentBranch()
	ProcessSymlinks()
	RepoPathExists() bool
	SetDefaultRef()
	SetRepoFullPath()
	SetRepoLocalName()
	SetSha()
}

// Set in place default name of the ref to master if not specified
func (repo *GitGetRepo) SetDefaultRef() {
	if repo.Ref == "" {
		repo.Ref = "master"
	}
}

func (repo *GitGetRepo) EnsurePathExists() {
	// if path is not specified set it to current working directory and return
	wdir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	if repo.Path == "" {
		repo.Path = wdir
		return
	}

	// otherwise ensure it is created
	repo.Path = path.Join(wdir, repo.Path)
	err = os.MkdirAll(repo.Path, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func (repo *GitGetRepo) SetRepoLocalName() {
	repo.AltName = repo.GetRepoLocalName()
}

func (repo *GitGetRepo) SetSha() {
	h := sha1.New()
	io.WriteString(h, fmt.Sprintf("%s (%s) %s", repo.Url, repo.Ref, repo.fullPath))
	repo.sha = fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

func (repo *GitGetRepo) Prepare() {
	repo.EnsurePathExists()
	repo.SetDefaultRef()
	repo.SetRepoLocalName()
	repo.SetRepoFullPath()
	repo.SetSha()

	log.Infof("%s: url: %s (%s) -> %s", repo.sha, repo.Url, colorRef.Sprintf(repo.Ref), repo.fullPath)
	log.Debugf("%s: path: '%s'", repo.sha, repo.Path)
	log.Debugf("%s: ref: '%s'", repo.sha, colorRef.Sprintf(repo.Ref))
	log.Debugf("%s: altName: '%s'", repo.sha, repo.AltName)
}

func (repo GitGetRepo) GetRepoLocalName() string {
	if repo.AltName == "" {
		re := regexp.MustCompile(`.*/`)
		repoName := re.ReplaceAllString(repo.Url, "")

		// remove trailing .git in repo name
		re = regexp.MustCompile(`.git$`)
		return re.ReplaceAllString(repoName, "")
	}
	return repo.AltName
}

func (repo *GitGetRepo) SetRepoFullPath() {
	repo.fullPath = path.Join(repo.Path, repo.GetRepoLocalName())
}

func PathExists(path string) (bool, os.FileInfo) {
	finfo, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, finfo
	}
	return true, finfo
}

func (repo GitGetRepo) RepoPathExists() bool {
	res, _ := PathExists(repo.fullPath)
	return res
}

func (repo GitGetRepo) Clone() bool {
	log.Infof("%s: Clone repository '%s'", repo.sha, repo.Url)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand(
		[]string{"clone", "--branch", repo.Ref, repo.Url, repo.fullPath},
		nil,
		&serr,
		"",
	)
	if err != nil {
		log.Errorf("%s: %v %v", repo.sha, err, serr.String())
		return false
	}
	return true
}

func (repo GitGetRepo) ShallowClone() bool {
	log.Infof("%s: Clone repository '%s'", repo.sha, repo.Url)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand(
		[]string{"clone", "--depth", "1", "--branch", repo.Ref, repo.Url, repo.fullPath},
		nil,
		&serr,
		"",
	)
	if err != nil {
		log.Errorf("%s: %v %v", repo.sha, err, serr.String())
		return false
	}
	return true
}

func (repo GitGetRepo) RemoveTargetDir(dotGit bool) {
	pathToRemove := ""
	if dotGit {
		pathToRemove = filepath.Join(repo.fullPath, ".git")
	} else {
		pathToRemove = repo.fullPath
	}
	err := os.RemoveAll(pathToRemove)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func (repo GitGetRepo) IsClean() bool {
	res := true
	_, err := repo.ExecGitCommand([]string{"diff", "--quiet"}, nil, nil, repo.fullPath)
	if err != nil {
		res = false
	}
	_, err = repo.ExecGitCommand([]string{"diff", "--staged", "--quiet"}, nil, nil, repo.fullPath)
	if err != nil {
		res = false
	}
	return res
}

func (repo GitGetRepo) IsCurrentBranchRef() bool {
	var outb, errb bytes.Buffer
	repo.ExecGitCommand([]string{"rev-parse", "--abbrev-ref", "HEAD"}, &outb, &errb, repo.fullPath)
	return (strings.TrimSpace(outb.String()) == repo.Ref)
}

func (repo GitGetRepo) GetCurrentBranch() string {
	var outb, errb bytes.Buffer
	repo.ExecGitCommand([]string{"rev-parse", "--abbrev-ref", "HEAD"}, &outb, &errb, repo.fullPath)
	return strings.TrimSpace(outb.String())
}

func (repo GitGetRepo) ExecGitCommand(
	args []string,
	stdoutb *bytes.Buffer,
	erroutb *bytes.Buffer,
	dir string,
) (cmd *exec.Cmd, err error) {
	cmd = exec.Command(gitCmd, args...)

	if stdoutb != nil {
		cmd.Stdout = stdoutb
	}
	if erroutb != nil {
		cmd.Stderr = erroutb
	}

	if dir != "" {
		cmd.Dir = dir
	}

	err = cmd.Run()
	return cmd, err
}

func (repo GitGetRepo) GitStashSave() {
	log.Infof("%s: Stash unsaved changes", repo.sha)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"stash", "save"}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Warnf("%s: %v: %v", repo.sha, err, serr.String())
	}
}

func (repo GitGetRepo) GitStashPop() {
	log.Infof("%s: Restore stashed changes", repo.sha)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"stash", "pop"}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Warnf("%s: %v: %v", repo.sha, err, serr.String())
	}
}

func (repo GitGetRepo) IsRefBranch() bool {
	res := true
	fullRef := fmt.Sprintf("refs/heads/%s", repo.Ref)
	_, err := repo.ExecGitCommand([]string{"show-ref", "--quiet", "--verify", fullRef}, nil, nil, repo.fullPath)
	if err != nil {
		res = false
	}
	return res
}

func (repo GitGetRepo) IsRefTag() bool {
	res := true
	fullRef := fmt.Sprintf("refs/tags/%s", repo.Ref)
	_, err := repo.ExecGitCommand([]string{"show-ref", "--quiet", "--verify", fullRef}, nil, nil, repo.fullPath)
	if err != nil {
		res = false
	}
	return res
}

func (repo GitGetRepo) GitPull() {
	if repo.IsRefBranch() {
		log.Infof("%s: Pulling upstream changes", repo.sha)
		var serr bytes.Buffer
		_, err := repo.ExecGitCommand([]string{"pull", "-f"}, nil, &serr, repo.fullPath)
		if err != nil {
			log.Errorf("%s: %v: %v", repo.sha, err, serr.String())
		}
	} else {
		log.Debugf("%s: Skip pulling upstream changes for '%s' which is not a branch", repo.sha, colorRef.Sprintf(repo.Ref))
	}
}

func (repo *GitGetRepo) ProcessRepoBasedOnCleaness() {
	if repo.IsClean() {
		log.Debugf("%s: Repo Status is clean", repo.sha)
		repo.GitPull()
	} else {
		log.Debugf("%s: Repo is NOT clean", repo.sha)
		repo.GitStashSave()

		repo.GitPull()

		repo.GitStashPop()
	}
}

func (repo GitGetRepo) GitCheckout(branch string) {
	log.Infof("%s: Checkout to '%s' branch in '%s'", repo.sha, colorHighlight.Sprintf(branch), repo.fullPath)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"checkout", branch}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Warnf("%s: %v: %v", repo.sha, err, serr.String())
	}
}

func (repo *GitGetRepo) ProcessRepoBasedOnCurrentBranch() {
	if repo.IsCurrentBranchRef() {
		log.Debugf("%s: Current branch is ref", repo.sha)
		repo.ProcessRepoBasedOnCleaness()
	} else {
		log.Debugf("%s: Current branch is not ref", repo.sha)
		currentBranch := repo.GetCurrentBranch()

		repo.GitCheckout(repo.Ref)
		repo.ProcessRepoBasedOnCleaness()

		if !stayOnRef {
			repo.GitCheckout(currentBranch)
		} else {
			log.Debugf("%s: Stay on ref branch '%s'", repo.sha, colorRef.Sprintf(repo.Ref))
		}
	}
}

func (repo GitGetRepo) CreateSymlink(symlink string) {
	log.Infof("%s: Processing symlink", repo.sha)
	// Check if exists - return
	exists, _ := PathExists(symlink)
	if exists {
		log.Debugf("%s: path for symlink '%s' exists (may not be symlink, don't care)", repo.sha, symlink)
		return
	}

	// check if directory of symlink exists
	symnlinkDir := filepath.Dir(symlink)
	exists, finfo := PathExists(symnlinkDir)
	if exists {
		// create symlink in it if it does
		if finfo.IsDir() {
			os.Symlink(repo.fullPath, symlink)
		} else {
			log.Errorf("%s: path for symlink '%s' directory '%s' exists, but is not directory - check configuration", repo.sha, symlink, symnlinkDir)
		}
	} else {
		// Otherwise ensure directory and create symlink
		err := os.MkdirAll(symnlinkDir, os.ModePerm)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		os.Symlink(repo.fullPath, symlink)
	}
}

func (repo GitGetRepo) ProcessSymlinks() {
	for _, symlink := range repo.Symlinks {
		repo.CreateSymlink(symlink)
	}
}

func processConfigShallow(repoList []GitGetRepo, concurrencyLevel int) {
	var throttle = make(chan int, concurrencyLevel)

	var wait sync.WaitGroup

	for _, repo := range repoList {
		throttle <- 1
		wait.Add(1)
		go func(repository GitGetRepo, iwait *sync.WaitGroup, ithrottle chan int) {
			defer iwait.Done()

			repository.Prepare()
			if repository.RepoPathExists() {
				log.Debugf("%s: path '%s' exists - removing target path", repository.sha, repository.fullPath)
				repository.RemoveTargetDir(false)
			}
			log.Debugf("%s: path '%s' missing - performing shallow clone", repository.sha, repository.fullPath)
			repository.ShallowClone()
			repository.RemoveTargetDir(true)
			repository.ProcessSymlinks()

			<-ithrottle
		}(repo, &wait, throttle)
	}

	wait.Wait()
}

func processConfigParallel(repoList []GitGetRepo, concurrencyLevel int) {
	var throttle = make(chan int, concurrencyLevel)

	var wait sync.WaitGroup

	for _, repo := range repoList {
		throttle <- 1
		wait.Add(1)

		go func(repository GitGetRepo, iwait *sync.WaitGroup, ithrottle chan int) {
			defer iwait.Done()

			repository.Prepare()
			if !repository.RepoPathExists() {
				// Clone
				log.Debugf("%s: path '%s' missing - cloning", repository.sha, repository.fullPath)
				repository.Clone()
			} else {
				// Refresh
				log.Debugf("%s: path '%s' exists, will refresh from remote", repository.sha, repository.fullPath)
				repository.ProcessRepoBasedOnCurrentBranch()
			}
			repository.ProcessSymlinks()

			<-ithrottle
		}(repo, &wait, throttle)
	}

	wait.Wait()
}

func processConfig(repoList []GitGetRepo) {
	for _, repository := range repoList {
		repository.Prepare()
		if !repository.RepoPathExists() {
			// Clone
			log.Debugf("%s: path '%s' missing - cloning", repository.sha, repository.fullPath)
			repository.Clone()
		} else {
			// Refresh
			log.Debugf("%s: path '%s' exists, will refresh from remote", repository.sha, repository.fullPath)
			repository.ProcessRepoBasedOnCurrentBranch()
		}
		repository.ProcessSymlinks()
	}
}

func GitGetRepositories(cfgFile string, concurrencyLevel int, stickToRef bool, shallow bool) {
	colorHighlight = color.New(color.FgRed)
	colorRef = color.New(color.FgHiBlue)
	yamlFile, err := ioutil.ReadFile(cfgFile)

	stayOnRef = stickToRef
	if err != nil {
		log.Fatalf("%s: %s", cfgFile, err)
	}

	var repoList []GitGetRepo
	if err := yaml.Unmarshal(yamlFile, &repoList); err != nil {
		log.Fatalf("%s: %s", cfgFile, err)
	}

	if shallow {
		processConfigShallow(repoList, concurrencyLevel)
	} else {
		processConfigParallel(repoList, concurrencyLevel)
	}
}
