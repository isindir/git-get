![](https://github.com/isindir/git-get/workflows/build-git-get/badge.svg)

# git-get - fetch multiple repositories

# Configuration file `Gitfile`

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

```
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

Available Commands:
  help        Help about any command
  version     Prints version information

Flags:
  -c, --concurrency-level int   Git get concurrnecy level (default 1)
  -f, --config-file string      configuration file (default "$PWD/Gitfile")
  -h, --help                    help for git-get
  -l, --log-level string        Logging level [debug|info|warn|error|fatal|panic] (default "info")
  -s, --shallow                 Shallow clone, can be used in CI to fetch dependencies by ref
  -t, --stay-on-ref             After refreshing repository from remote stay on ref branch

Use "git-get [command] --help" for more information about a command.
```
