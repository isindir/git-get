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

// Package gitget implements business logic of the application.
package gitget

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"golang.org/x/net/context"
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

	"github.com/isindir/git-get/bitbucket"
	"github.com/isindir/git-get/github"
	"github.com/isindir/git-get/gitlab"
)

const gitCmd = "git"

var stayOnRef bool
var defaultMainBranch = "master"
var mirrorProvider string
var mirrorVisibilityMode = "private"
var bitbucketMirrorProject = ""
var colorHighlight *color.Color
var colorRef *color.Color

// Repo structure defines information about single git repository.
type Repo struct {
	URL       string   `yaml:"url"`
	Path      string   `yaml:"path,omitempty"`
	AltName   string   `yaml:"altname,omitempty"`
	Ref       string   `yaml:"ref,omitempty"`
	Symlinks  []string `yaml:"symlinks,omitempty"`
	fullPath  string
	sha       string
	mirrorURL string
}

// RepoI interface defined for mocking purposes.
type RepoI interface {
	Clone() bool
	CreateSymlink(symlink string)
	ChoosePathPrefix(pathPrefix string) string
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
	PrepareForGet()
	ProcessRepoBasedOnCleaness()
	ProcessRepoBasedOnCurrentBranch()
	ProcessSymlinks()
	RepoPathExists() bool
	SetDefaultRef()
	SetRepoFullPath()
	SetRepoLocalName()
	SetSha()
}

func initColors() {
	colorHighlight = color.New(color.FgRed)
	colorRef = color.New(color.FgHiBlue)
}

// SetDefaultRef sets in place default name of the ref to master (by default) or user passed via flag if not specified
func (repo *Repo) SetDefaultRef() {
	if repo.Ref == "" {
		repo.Ref = defaultMainBranch
	}
}

func (repo *Repo) ChoosePathPrefix(pathPrefix string) string {
	if pathPrefix != "" {
		// If pathPrefix does not exist or is not Directory - fail
		exists, fileInfo := PathExists(pathPrefix)
		if exists && fileInfo.IsDir() {
			return pathPrefix
		}
		log.Fatalf("Error: %s does not exist or is not directory", pathPrefix)
		os.Exit(1)
	} else {
		workingDirectory, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		return workingDirectory
	}

	return ""
}

func (repo *Repo) SetTempRepoPathForMirror(pathPrefix string) {
	repo.Path = repo.ChoosePathPrefix(pathPrefix)
}

func (repo *Repo) EnsurePathExists(pathPrefix string) {
	// if path is not specified set it to current working directory, otherwise use passed value
	wdir := repo.ChoosePathPrefix(pathPrefix)

	if repo.Path == "" {
		repo.Path = wdir
		return
	}

	// otherwise, ensure it is created
	repo.Path = path.Join(wdir, repo.Path)
	err := os.MkdirAll(repo.Path, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

func (repo *Repo) SetRepoLocalName() {
	repo.AltName = repo.GetRepoLocalName()
}

// SetSha generates and sets `sha` of the data structure to use in log messages.
func (repo *Repo) SetSha() {
	h := sha1.New()
	io.WriteString(h, fmt.Sprintf("%s (%s) %s", repo.URL, repo.Ref, repo.fullPath))
	repo.sha = fmt.Sprintf("%x", h.Sum(nil))[0:7]
}

// PrepareForGet performs checks for repository as well as constructs extra information and sets repository data structure
// values.
func (repo *Repo) PrepareForGet() {
	repo.EnsurePathExists("")
	repo.SetDefaultRef()
	repo.SetRepoLocalName()
	repo.SetRepoFullPath()
	repo.SetSha()

	log.Infof("%s: url: %s (%s) -> %s", repo.sha, repo.URL, colorRef.Sprintf(repo.Ref), repo.fullPath)
	log.Debugf("%s: Repository structure: '%+v'", repo.sha, repo)
}

func (repo *Repo) PrepareForMirror(pathPrefix string, mirrorRootURL string) {
	repo.SetTempRepoPathForMirror(pathPrefix)
	repo.SetDefaultRef()
	repo.SetRepoLocalName()
	repo.SetMirrorURL(mirrorRootURL)
	repo.SetRepoFullPath()
	repo.SetSha()

	log.Infof("%s: url: %s (%s) -> %s", repo.sha, repo.URL, colorRef.Sprintf(repo.Ref), repo.fullPath)
	log.Debugf("%s: Repository structure: '%+v'", repo.sha, repo)
}

func (repo Repo) GetRepoLocalName() string {
	if repo.AltName == "" {
		re := regexp.MustCompile(`.*/`)
		repoName := re.ReplaceAllString(repo.URL, "")

		// remove trailing .git in repo name
		re = regexp.MustCompile(`.git$`)
		return re.ReplaceAllString(repoName, "")
	}
	return repo.AltName
}

func (repo *Repo) SetMirrorURL(mirrorRootURL string) {
	repo.mirrorURL = fmt.Sprintf("%s/%s.%s", mirrorRootURL, repo.AltName, "git")
}

func (repo *Repo) SetRepoFullPath() {
	repo.fullPath = path.Join(repo.Path, repo.GetRepoLocalName())
}

// PathExists returns `true` if given `path` passed as sting exists, otherwise returns false.
func PathExists(path string) (bool, os.FileInfo) {
	finfo, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, finfo
	}
	return true, finfo
}

func (repo Repo) RepoPathExists() bool {
	res, _ := PathExists(repo.fullPath)
	return res
}

// CloneMirror runs `git clone --mirror` command.
func (repo Repo) CloneMirror() bool {
	log.Infof("%s: Clone repository '%s' for mirror", repo.sha, repo.URL)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand(
		[]string{"clone", "--mirror", repo.URL, repo.fullPath},
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

// PushMirror runs `git push --mirror` command.
func (repo Repo) PushMirror() bool {
	log.Infof("%s: Push repository '%s' as a mirror '%s'", repo.sha, repo.URL, repo.mirrorURL)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"push", "--mirror", repo.mirrorURL}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Errorf("%s: %v %v", repo.sha, err, serr.String())
		return false
	}
	return true
}

// Clone runs `git clone --branch` command.
func (repo Repo) Clone() bool {
	log.Infof("%s: Clone repository '%s'", repo.sha, repo.URL)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand(
		[]string{"clone", "--branch", repo.Ref, repo.URL, repo.fullPath},
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

// ShallowClone runs `git clone --depth 1 --branch` command.
func (repo Repo) ShallowClone() bool {
	log.Infof("%s: Clone repository '%s'", repo.sha, repo.URL)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand(
		[]string{"clone", "--depth", "1", "--branch", repo.Ref, repo.URL, repo.fullPath},
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

func (repo Repo) RemoveTargetDir(dotGit bool) {
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

func (repo Repo) IsClean() bool {
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

func (repo Repo) IsCurrentBranchRef() bool {
	var outb, errb bytes.Buffer
	repo.ExecGitCommand([]string{"rev-parse", "--abbrev-ref", "HEAD"}, &outb, &errb, repo.fullPath)
	return (strings.TrimSpace(outb.String()) == repo.Ref)
}

func (repo Repo) GetCurrentBranch() string {
	var outb, errb bytes.Buffer
	repo.ExecGitCommand([]string{"rev-parse", "--abbrev-ref", "HEAD"}, &outb, &errb, repo.fullPath)
	return strings.TrimSpace(outb.String())
}

func (repo Repo) ExecGitCommand(
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

func (repo Repo) GitStashSave() {
	log.Infof("%s: Stash unsaved changes", repo.sha)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"stash", "save"}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Warnf("%s: %v: %v", repo.sha, err, serr.String())
	}
}

func (repo Repo) GitStashPop() {
	log.Infof("%s: Restore stashed changes", repo.sha)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"stash", "pop"}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Warnf("%s: %v: %v", repo.sha, err, serr.String())
	}
}

func (repo Repo) IsRefBranch() bool {
	res := true
	fullRef := fmt.Sprintf("refs/heads/%s", repo.Ref)
	_, err := repo.ExecGitCommand([]string{"show-ref", "--quiet", "--verify", fullRef}, nil, nil, repo.fullPath)
	if err != nil {
		res = false
	}
	return res
}

func (repo Repo) IsRefTag() bool {
	res := true
	fullRef := fmt.Sprintf("refs/tags/%s", repo.Ref)
	_, err := repo.ExecGitCommand([]string{"show-ref", "--quiet", "--verify", fullRef}, nil, nil, repo.fullPath)
	if err != nil {
		res = false
	}
	return res
}

func (repo Repo) GitPull() {
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

func (repo *Repo) ProcessRepoBasedOnCleaness() {
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

func (repo Repo) GitCheckout(branch string) {
	log.Infof("%s: Checkout to '%s' branch in '%s'", repo.sha, colorHighlight.Sprintf(branch), repo.fullPath)
	var serr bytes.Buffer
	_, err := repo.ExecGitCommand([]string{"checkout", branch}, nil, &serr, repo.fullPath)
	if err != nil {
		log.Warnf("%s: %v: %v", repo.sha, err, serr.String())
	}
}

func (repo *Repo) ProcessRepoBasedOnCurrentBranch() {
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

func (repo Repo) CreateSymlink(symlink string) {
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
		// create symlink in directory if it does exist
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

func (repo Repo) ProcessSymlinks() {
	for _, symlink := range repo.Symlinks {
		repo.CreateSymlink(symlink)
	}
}

func (repo *Repo) EnsureMirrorExists() {
	switch mirrorProvider {
	case "gitlab":
		repo.EnsureGitlabMirrorExists()
	case "github":
		repo.EnsureGithubMirrorExists()
	case "bitbucket":
		repo.EnsureBitbucketMirrorExists()
	default:
		log.Fatalf("%s: Error: unknown '%s' git mirror provider", repo.sha, mirrorProvider)
		os.Exit(1)
	}
}

func (repo *Repo) EnsureGitlabMirrorExists() {
	// ( a/b/c/d -> a , b , b/c/d, d )
	baseURL, projectNameFullPath, projectNameShort := repo.DecomposeGitURL()
	log.Debugf("%s: For Check: BaseURL: %s projectNameFullPath: %s", repo.sha, baseURL, projectNameFullPath)
	log.Debugf("%s: For Create: BaseURL: %s projectNameShort: %s", repo.sha, baseURL, projectNameShort)
	// In gitlab Project is both - repository and directory to aggregate repositories
	projectFound := gitlab.ProjectExists(repo.sha, baseURL, projectNameFullPath)

	if !projectFound {
		log.Debugf("%s: Gitlab project '%s' does not exist", repo.sha, projectNameFullPath)
		// identify if part `b` is a group ? then need to create project differently - potentially create all subgroups
		projectNamespace, namespaceFullPath := gitlab.GetProjectNamespace(repo.sha, baseURL, projectNameFullPath)
		log.Debugf("%s: '%s' is '%+v'", repo.sha, projectNameFullPath, projectNamespace)
		// If project namespace exists and is user - create project for user
		// Otherwise ensure group with subgroups exists and create project in subgroup
		if projectNamespace != nil {
			if projectNamespace.Kind == "user" {
				log.Debugf("%s: Creating new gitlab project '%s' on '%s' for user '%s'", repo.sha, projectNameShort, baseURL, projectNamespace.Path)
				gitlab.CreateProject(repo.sha, baseURL, projectNameShort, 0, mirrorVisibilityMode, repo.URL)
			} else {
				log.Debugf("%s: Creating new gitlab project '%s' on '%s' for namespace '%s'", repo.sha, projectNameShort, baseURL, projectNamespace.Path)
				gitlab.CreateProject(repo.sha, baseURL, projectNameShort, projectNamespace.ID, mirrorVisibilityMode, repo.URL)
			}
		} else {
			// TODO: space for improvement - ensure group with subgroups is created, then create project
			log.Fatalf("%s: Group '%s' does not exist, please ensure it is created using for mirrors", repo.sha, namespaceFullPath)
			os.Exit(1)
		}
	}
}

func (repo *Repo) DecomposeGitURL() (string, string, string) {
	// remove unwanted parts of the git repo url
	re := regexp.MustCompile(`.git$|^https://|^git@`)
	url := re.ReplaceAllString(repo.mirrorURL, "")
	re = regexp.MustCompile(`:`)
	url = re.ReplaceAllString(url, "/")

	// baseURL and longPath for checking repo existence ( a/b/c/d -> a , b/c/d )
	urlParts := strings.SplitN(url, "/", 2)
	baseURL, fullName := urlParts[0], urlParts[1]

	// baseURL and projecName for creating missing repository ( a/b/c/d -> a/b/c , d)
	_, shortName := filepath.Split(url)

	// ( a/b/c/d -> a, b, b/c/d )
	return baseURL, fullName, shortName
}

func (repo *Repo) EnsureGithubMirrorExists() {
	_, projectNameFullPath, _ := repo.DecomposeGitURL()
	repoNameParts := strings.SplitN(projectNameFullPath, "/", 2)
	workspaceName, repositoryName := repoNameParts[0], repoNameParts[1]
	ctx := context.Background()
	if !github.RepositoryExists(ctx, repo.sha, workspaceName, repositoryName) {
		log.Debugf("%s: Creating new github repository '%s'", repo.sha, repo.mirrorURL)
		github.CreateRepository(ctx, repo.sha, repositoryName, mirrorVisibilityMode, repo.URL)
	} else {
		log.Debugf("%s: github repository '%s' exists", repo.sha, repo.mirrorURL)
	}
}

func (repo *Repo) EnsureBitbucketMirrorExists() {
	_, fullName, _ := repo.DecomposeGitURL()
	repoNameParts := strings.SplitN(fullName, "/", 2)
	workspaceName, repositoryName := repoNameParts[0], repoNameParts[1]
	if !bitbucket.RepositoryExists(repo.sha, workspaceName, repositoryName) {
		log.Debugf("%s: Creating new bitbucket repository '%s'", repo.sha, repo.mirrorURL)
		bitbucket.CreateRepository(repo.sha, fullName, mirrorVisibilityMode, repo.URL, bitbucketMirrorProject)
	} else {
		log.Debugf("%s: bitbucket repository '%s' exists", repo.sha, repo.mirrorURL)
	}
}

func getShallowReposFromConfigInParallel(repoList []Repo, concurrencyLevel int) {
	var throttle = make(chan int, concurrencyLevel)

	var wait sync.WaitGroup

	for _, repo := range repoList {
		throttle <- 1
		wait.Add(1)
		go func(repository Repo, iwait *sync.WaitGroup, ithrottle chan int) {
			defer iwait.Done()

			repository.PrepareForGet()
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

func getReposFromConfigInParallel(repoList []Repo, concurrencyLevel int) {
	var throttle = make(chan int, concurrencyLevel)

	var wait sync.WaitGroup

	for _, repo := range repoList {
		throttle <- 1
		wait.Add(1)

		go func(repository Repo, iwait *sync.WaitGroup, ithrottle chan int) {
			defer iwait.Done()

			repository.PrepareForGet()
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

func mirrorReposFromConfigInParallel(repoList []Repo, concurrencyLevel int, pushMirror bool, mirrorRootURL string) {
	var throttle = make(chan int, concurrencyLevel)

	var wait sync.WaitGroup

	// make temp directory - preserve it's name
	tempDir, err := ioutil.TempDir("", "gitgetmirror")
	if err != nil {
		log.Fatalf("Error: %s, while creating temporary directory", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	for _, repo := range repoList {
		throttle <- 1
		wait.Add(1)

		go func(repository Repo, iwait *sync.WaitGroup, ithrottle chan int) {
			defer iwait.Done()

			repository.PrepareForMirror(tempDir, mirrorRootURL)
			// Clone
			log.Debugf("%s: path '%s' cloning for mirror", repository.sha, repository.fullPath)
			repository.CloneMirror()
			if pushMirror {
				repository.EnsureMirrorExists()
				repository.PushMirror()
			} else {
				log.Infof("%s: skipping '%s' remote push per user request", repository.sha, repository.URL)
			}

			<-ithrottle
		}(repo, &wait, throttle)
	}

	wait.Wait()
}

func processConfig(repoList []Repo) {
	for _, repository := range repoList {
		repository.PrepareForGet()
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

func GetRepositories(cfgFile string, concurrencyLevel int, stickToRef bool, shallow bool, defaultTrunkBranch string) {
	initColors()
	stayOnRef = stickToRef
	defaultMainBranch = defaultTrunkBranch

	yamlFile, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Fatalf("%s: %s", cfgFile, err)
	}

	var repoList []Repo
	if err := yaml.Unmarshal(yamlFile, &repoList); err != nil {
		log.Fatalf("%s: %s", cfgFile, err)
	}

	if shallow {
		getShallowReposFromConfigInParallel(repoList, concurrencyLevel)
	} else {
		getReposFromConfigInParallel(repoList, concurrencyLevel)
	}
}

func MirrorRepositories(
	cfgFile string,
	concurrencyLevel int,
	pushMirror bool,
	mirrorRootURL string,
	mirrorProviderName string,
	mirrorVisibilityModeName string,
	mirrorBitbucketProjectName string,
) {
	initColors()
	mirrorProvider = mirrorProviderName
	mirrorVisibilityMode = mirrorVisibilityModeName
	bitbucketMirrorProject = mirrorBitbucketProjectName

	yamlFile, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Fatalf("%s: %s", cfgFile, err)
	}

	var repoList []Repo
	if err := yaml.Unmarshal(yamlFile, &repoList); err != nil {
		log.Fatalf("%s: %s", cfgFile, err)
	}

	mirrorReposFromConfigInParallel(repoList, concurrencyLevel, pushMirror, mirrorRootURL)
}
