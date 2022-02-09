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

`git-get` (which can be executed as `git get`) fetches git repositories using
configuration file (by default `Gitfile` in current directory of invocation).
`Gitfile.ignore` file can control which repositories to skip from the operations
and it has same format, where in `Gitfile.ignore` the only significant field is
`url`, which value is compared with the target one. If `Gitfile.ignore` is missing
this functionality is ignored.

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

## Other `git-get` operations

`git-get` can also generate configuration file from repositories in git provider or mirror
repositories specified by `Gitfile` to a chosen git provider. For the clone/fetch operations
on git repositories ssh keys must be used. For creating target repositories in git provider
(during mirror operation) or for fetching the list of available repositories in git provider -
user API keys are used. `git-get` allows shallow clone of the repositories, which is suitable
for use in CI/CD.

# Command line options

## Fetching/Refreshing repositories specified by Gitfile

```bash
% git-get --help
'git-get' - all your project repositories

git-get clone/refresh all your local project repositories in
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
  config-gen  Create Gitfile configuration file from git provider
  help        Help about any command
  mirror      Create or update repositories mirror in a specified git provider cloud
  version     Prints version information

Flags:
  -c, --concurrency-level int        Git get concurrency level (default 1)
  -f, --config-file string           Configuration file (default "~/Gitfile")
  -b, --default-main-branch string   Default main branch (default "master")
  -h, --help                         help for git-get
  -i, --ignore-file string           Ignore file (default "~/Gitfile.ignore")
  -l, --log-level string             Logging level [debug|info|warn|error|fatal|panic] (default "info")
  -s, --shallow                      Shallow clone, can be used in CI to fetch dependencies by ref
      --status                       Print extra status information after clone is performed
  -t, --stay-on-ref                  After refreshing repository from remote stay on ref branch

Use "git-get [command] --help" for more information about a command.
```

## Generating Gitfile from git provider

```bash
% git-get config-gen --help

Create 'Gitfile' configuration file dynamically from git provider by specifying
top level URL of the organisation organization, user or for Gitlab provider Group name.

* Github: Environment variable GITHUB_TOKEN defined.
* Bitbucket: Environment variables BITBUCKET_USERNAME and BITBUCKET_TOKEN (password) defined.
* Gitlab: Environment variable GITLAB_TOKEN defined.
* Gitlab: provider allows to create hierarchy of groups, 'git-get' is capable of fetching
  this hierarchy to 'Gifile' from any level visible to the user (see examples).

Usage:
  git-get config-gen [flags]

Examples:

git-get config-gen -f Gitfile -p "gitlab" -u "git@gitlab.com:johndoe" -t misc -l debug
git-get config-gen -f Gitfile -p "gitlab" -u "git@gitlab.com:AcmeOrg" -t misc -l debug
git-get config-gen -f Gitfile -p "gitlab" -u "git@gitlab.com:AcmeOrg/kube"
git-get config-gen -f Gitfile -p "bitbucket" -u "git@bitbucket.com:AcmeOrg" -t AcmeOrg
git-get config-gen -f Gitfile -p "github" -u "git@github.com:johndoe" -t johndoe -l debug
git-get config-gen -f Gitfile -p "github" -u "git@github.com:AcmeOrg" -t AcmeOrg -l debug

Flags:
      --bitbucket-role string                       Bitbucket: Filter repositories by role [owner|admin|contributor|member] (default "member")
  -f, --config-file string                          Configuration file (default "~/Gitfile")
  -p, --config-provider string                      Git provider name [gitlab|github|bitbucket] (default "gitlab")
  -u, --config-url string                           Private URL prefix to construct Gitfile from (example: git@github.com:acmeorg), provider specific.
      --github-affiliation string                   Github: affiliation - comma-separated list of values.
                                                    Can include: owner, collaborator, or organization_member (default "owner,collaborator,organization_member")
      --github-visibility string                    Github: visibility [all|public|private] (default "all")
      --gitlab-groups-minimal-access-level string   Gitlab: groups minimal access level [unspecified|min|guest|reporter|developer|maintainer|owner] (default "unspecified")
      --gitlab-owned                                Gitlab: only traverse groups and repositories owned by user
      --gitlab-project-visibility string            Gitlab: project visibility [public|internal|private]
  -h, --help                                        help for config-gen
  -i, --ignore-file string                          Ignore file (default "~/Gitfile.ignore")
  -l, --log-level string                            Logging level [debug|info|warn|error|fatal|panic] (default "info")
  -t, --target-clone-path string                    Target clone path used to set 'path' for each repository in Gitfile
```

## Creating mirror repositories in git provider

```bash
git-get mirror --help

Create or update repositories mirror in a specified specified git provider cloud using configuration file.

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

git get mirror -f Gitfile -u "git@github.com:acmeorg" -p "github"
git-get mirror -c 2 -f Gitfile -l debug -u "git@gitlab.com:acmeorg/mirrors"
git-get mirror -c 2 -f Gitfile -l debug -u "git@bitbucket.com:acmeorg" -p "bitbucket" -b "mirrors"

Flags:
  -b, --bitbucket-mirror-project-name string   Bitbucket mirror project name (only effective for Bitbucket and is optional)
  -c, --concurrency-level int                  Git get concurrency level (default 1)
  -f, --config-file string                     Configuration file (default "~/Gitfile")
  -d, --dry-run                                Dry-run - do not push to remote mirror repositories
  -h, --help                                   help for mirror
  -i, --ignore-file string                     Ignore file (default "~/Gitfile.ignore")
  -l, --log-level string                       Logging level [debug|info|warn|error|fatal|panic] (default "info")
  -p, --mirror-provider string                 Git mirror provider name [gitlab|github|bitbucket] (default "gitlab")
  -u, --mirror-url string                      Private Mirror URL prefix to push repositories to (example: git@github.com:acmeorg)
  -v, --mirror-visibility-mode string          Mirror visibility mode [private|internal|public] (default "private")
```

# Related or similar projects

* https://github.com/bradurani/Gitfile
* https://github.com/coretech/terrafile
* https://github.com/grdl/git-get
* https://github.com/x-motemen/ghq
* https://github.com/fboender/multi-git-status
