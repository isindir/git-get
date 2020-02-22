* aggregate log entries for each item - to print like these would be run sequentially

```
- url: git@github.com:isindir/git-get.git
- url: git@github.com:isindir/git-get.git
  path: DevOps/deployments/
  altname: qqq
- url: git@github.com:isindir/git-get.git
  path: DevOps/deployments/
  trunk: develop
- url: git@github.com:isindir/git-get.git
  path: ./DevOps/edit-conf
- url: git@github.com:isindir/git-get.git
  trunk: develop
  symlink: abc/cde
- url: git@github.com:isindir/git-get.git
  path: DevOps/deployments/abc
  symlink: qqq/cde
```
