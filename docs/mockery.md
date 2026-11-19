# Regenerating mocks (mockery)

Internal notes for maintainers on how to regenerate the testify mocks in
`exec/mocks`, `gitlab/mocks`, `github/mocks` and `bitbucket/mocks`.

## TL;DR

```bash
make tools                   # mise install: tools pinned in mise.toml, incl. mockery
make mockery                 # version check -> clean -> regenerate
go vet ./... && make test    # generated code must compile and tests must pass
git diff --stat -- '*/mocks' # review; commit mocks together with the change that caused them
```

## How it is wired

| Piece | Role |
|-------|------|
| `mise.toml` | Pins the mockery version (`mockery = "3.x.y"` under `[tools]`). Single source of truth. |
| `.mockery.yaml` | mockery v3 config: one entry per package, `all: true`, output to `<pkg>/mocks/mocks.go`, package `mocks`. |
| `MOCK_DIRS` (Makefile) | List of generated directories. Must match the `dir:` entries in `.mockery.yaml`. |
| `make mockery-version` | Fails if `mockery version` != `v<version from mise.toml>`. |
| `make clean-mockery` | `rm -fr $(MOCK_DIRS)`. |
| `make mockery` | `clean-mockery` + `mockery-version`, then runs `mockery` from the repo root. |
| `make mockery-verify` | `make mockery` then `git diff --exit-code` on `MOCK_DIRS`. Fails if committed mocks are stale. |

`make mod` depends on `make mockery`, so it also regenerates mocks.

## Why mocks drift

Generated output depends on the **mockery version**, not only on the
interfaces. Template changes between mockery releases change the files even
when no interface changed. Examples from moving 3.7.0 to 3.9.1:

- `interface{}` became `any` in `_Expecter` method signatures
- `t.Helper()` calls were added to the `New<Interface>` constructor and to its cleanup
- the closure parameter in `Run` was renamed (`args` to `args1`) to avoid clashing
  with method parameters named `args`

So **any change to the mockery version in `mise.toml` must be committed
together with regenerated mocks**. If it isn't, the next person who runs
`make mockery` gets an unrelated diff.

## Procedure

### When an interface changes

1. Edit the interface in `exec/`, `gitlab/`, `github/` or `bitbucket/`.
2. `make mockery`
3. `go vet ./... && make test`
4. Commit the interface change and the regenerated `*/mocks/mocks.go` together.

### When bumping mockery

1. Update `mockery` in `mise.toml`, then run `mise install mockery`.
2. `make mockery`. Expect a diff in all mock files that comes only from the template.
3. `go vet ./... && make test`
4. Commit `mise.toml` and the mock files together.

### When adding a package with interfaces to mock

1. Add a package entry to `.mockery.yaml` (copy an existing block and change
   the import path and `dir`).
2. Add the same `dir` to `MOCK_DIRS` in the `Makefile`.
3. `make mockery`

### Checking for stale mocks (e.g. before a PR)

```bash
make mockery-verify
```

It exits non-zero if the regenerated mocks differ from what is in git. It
compares against the working tree, so run it on a tree where the mock files
have no uncommitted changes.

## Troubleshooting

- **`mockery version mismatch`**: the `mockery` on `PATH` isn't the pinned
  version. Common causes are a Homebrew or `go install` binary shadowing the
  mise shim/activated path, mise not being activated in the shell, or the
  version not being installed. Check with `which -a mockery` and
  `mise which mockery`, then run `mise install mockery`.
- **mockery v2 binary**: v2 doesn't understand the v3 keys in `.mockery.yaml`
  (`structname`, `pkgname` and so on) and writes files in a different layout.
  Always use v3 as pinned.
- **`go vet`/`go build` fails after `make clean-mockery`**: expected.
  `gitget/get_test.go` imports `exec/mocks`. Run `make mockery` to restore it.
- **Unexpected diff after `make mockery` without any interface change**:
  the committed mocks were generated with a different mockery version. See
  [Why mocks drift](#why-mocks-drift). Commit the regenerated files.

## Known limitation

`gitlab.GitGetGitlabI` contains unexported methods (`gitlabAuth`,
`getGroupID`, `processSubgroups`, `appendGroupsProjects`). A mock generated
into the separate `mocks` package can't satisfy an interface with unexported
methods, so `gitlab/mocks.GitGetGitlabI` can't be used where a
`gitlab.GitGetGitlabI` is expected. Only `exec/mocks` is used by tests today
(`gitget/get_test.go`). To make the gitlab mock usable, move the unexported
methods out of the interface, or export them.
