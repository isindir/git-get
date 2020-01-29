* use yaml file for the directory structure of the project
  * list of lists - with 2 types of list items:
    * directory name
    * repository reference - `<full clone path>,<trunk branch>[,<directory name to clone to>]`
* parallel run to update/clone multiple repos at the same time (4: default, configurable)
* sequence of ops at run time:
  * detect - dir or repo
    * dir: exists?
      * yes: skip
      * no: create
    * repo:
      * no: clone
      * yes: on a trunk branch?
        * yes: has unstashed files ?
          * yes: stash
            * pull -f
            * unstash
          * no: pull -f
      * no: remember branch
        * yes: has unstashed files ?
          * yes: stash
            * checkout to trunk
            * pull -f
            * checkout to feature branch
            * unstash
          * no: checkout to trunk
            * pull -f
            * checkout to feature branch

```
- komgo:
  - DevOps:
    - git@gitlab.com:komgo/devops/editor-config.git,master
    - reports:
      - git@gitlab.com:komgo/devops/reports/aws-user-report.git,develop
      - git@gitlab.com:komgo/devops/reports/sonarreports.git,develop
    - network:
      - git@gitlab.com:komgo/devops/network/calico.git,develop
    - infrastructure:
      - git@gitlab.com:komgo/devops/infrastructure/helm.git,develop
```

```
- komgo:
  - name: DevOps
    type: dir
    items:
    - name: deployments
      type: dir
      items:
      - name: git@gitlab.com:komgo/devops/pipes.git
        type: repo
        trunk: develop
    - name: git@gitlab.com:komgo/devops/editor-config.git
      type: repo
      trunk: master
      altName: edit-conf
    - name: git@gitlab.com:komgo/devops/pipes.git
      type: repo
      trunk: develop
      symlink: $GOPATH/src/github.com/blabla/pipes
```

* clean code - almost one function per file, simple directory structure
* libs?
  * yaml processing
  * error handling
  * git operations (via library or git command ?)
