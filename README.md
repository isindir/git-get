[![Go Report Card](https://goreportcard.com/badge/github.com/isindir/git-get?)](https://goreportcard.com/report/github.com/isindir/git-get)
[![GoActions](https://github.com/isindir/git-get/workflows/build-git-get/badge.svg)](https://github.com/isindir/git-get/actions?query=workflow%3Abuild-git-get)
[![GitHub release](https://img.shields.io/github/tag/isindir/git-get.svg)](https://github.com/isindir/git-get/releases)
[![MIT](http://img.shields.io/github/license/isindir/git-get.svg)](LICENSE)

# git-get - fetch multiple repositories

## Installation

```bash
$ brew tap isindir/git-get
$ brew install git-get
```

## Configuration file `Gitfile`

`git-get` (which can be executed as `git get` if is in the `$PATH`) fetches git repositories using
configuration file (by default `Gitfile` in current directory of invocation).

`Gitfile` format is:

```
- url: git@github.com:isindir/git-get.git
  path: DevOps/deployments/
  ref: master
  altname: qqq1
  symlinks:
  - qqq/cde
  - qqq/edc
```

> all the fields, except `url:` are optional

## Configuration file fields

* `url` specifies repository to fetch
* `path` specifies relative to current directory path, where to clone git repository
* `ref` specifies the main branch of the repository which will be refreshed and optionally
  switched to
* `altname` specifies alternate name of the cloned repository, i.e. if repository name is `git-get` by
  specifying `altname: my-git-get` will clone repository into directory `my-git-get`
* `symlinks` is an optional list of paths to create symlinks to this clones repository. If such a file
  already exists (symlink, directory or regular file) - nothing will be done

# Command line options

## Fetching/Refreshing repositories specified by Gitfile

```bash
% git-get --help
'git-get' - all your project repositories

git-get clones/refreshes all you project repositories in
one go.

Yaml formatted configuration file specifies directory
structure of the project. git-get allows to create symlinks
to cloned repositories, clone one repository multiple time
having different directory name.

Usage:
  git-get [flags]
  git-get [command]
  
Examples:

git get -c 12 -f Gitfile

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  mirror      Creates or updates a mirror of repositories
  version     Prints version information

Flags:
  -c, --concurrency-level int        Git get concurrency level (default 1)
  -f, --config-file string           Configuration file (default "~/Gitfile")
  -b, --default-main-branch string   Default main branch (default "master")
  -h, --help                         help for git-get
  -l, --log-level string             Logging level [debug|info|warn|error|fatal|panic] (default "info")
  -s, --shallow                      Shallow clone, can be used in CI to fetch dependencies by ref
  -t, --stay-on-ref                  After refreshing repository from remote stay on ref branch

Use "git-get [command] --help" for more information about a command.
```

## Creating mirror repositories in git provider

```bash
git-get mirror --help

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
  only [A-Z0-9_] characters, all the rest of the characters from Project Name will be removed.

Usage:
  git-get mirror [flags]

Examples:

git get mirror -f Gitfile -u "git@github.com:acmeorg" -m "github"
git-get mirror -c 2 -f Gitfile -l debug -u "git@gitlab.com:acmeorg/mirrors"
git-get mirror -c 2 -f Gitfile -l debug -u "git@bitbucket.com:acmeorg" -m "bitbucket" -b "mirrors"

Flags:
  -b, --bitbucket-mirror-project-name string   Bitbucket mirror project name (only effective for Bitbucket and is optional)
  -c, --concurrency-level int                  Git get concurrency level (default 1)
  -f, --config-file string                     Configuration file (default "~/Gitfile")
  -h, --help                                   help for mirror
  -l, --log-level string                       Logging level [debug|info|warn|error|fatal|panic] (default "info")
  -m, --mirror-provider string                 Git mirror provider name [gitlab|github|bitbucket] (default "gitlab")
  -u, --mirror-url string                      Private Mirror URL prefix to push repositories to (example: git@github.com:acmeorg)
  -v, --mirror-visibility-mode string          Mirror visibility mode [private|internal|public] (default "private")
  -p, --push                                   Push to remote mirror repositories (default true)
```
