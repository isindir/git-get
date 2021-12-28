SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
GO := GOPROXY=https://proxy.golang.org go

VERSION:="0.0.7"
EXE:="git-get"
BUILD:=`git rev-parse --short HEAD`
TIME:=`date`
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: clean tidy fmt vet test build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: ## Builds the binary file
	@GO111MODULE=on GOOS=darwin GOARCH=amd64 $(GO) build \
		-ldflags="-X 'github.com/isindir/git-get/version.Version=v${VERSION}' \
		-X 'github.com/isindir/git-get/version.Commit=${BUILD}' \
		-X 'github.com/isindir/git-get/version.Time=${TIME}' " \
	  -o ./bin/${EXE}-${VERSION}-osx main.go
	@GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO) build \
		-ldflags="-X 'github.com/isindir/git-get/version.Version=v${VERSION}' \
		-X 'github.com/isindir/git-get/version.Commit=${BUILD}' \
		-X 'github.com/isindir/git-get/version.Time=${TIME}' " \
	  -o ./bin/${EXE}-${VERSION}-linux main.go
	@$(GO) build -ldflags="-X 'github.com/isindir/git-get/version.Version=v${VERSION}' \
		-X 'github.com/isindir/git-get/version.Commit=${BUILD}' \
		-X 'github.com/isindir/git-get/version.Time=${TIME}' " \
		-o ./bin/${EXE} main.go

.PHONY: run
run: ## Runs main help
	$(GO) run main.go

.PHONY: test
test: ## Placeholder to run unit tests
	@echo "Running unit tests"
	@mkdir -p bin
	$(GO) test -cover -coverprofile=bin/c.out ./...
	$(GO) tool cover -html=bin/c.out -o bin/coverage.html
	@echo

.PHONY: check
check: ## Runs linting
	@echo "Linting"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@echo

.PHONY: fmt
fmt: ## Run go fmt against code.
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	$(GO) vet ./...

.PHONY: clean
clean: ## Removes build artifacts from source code
	@echo "Cleaning"
	@rm -fr bin
	@rm -fr vendor
	@rm -fr chglog.tmp
	@echo

.PHONY: repo-tag
repo-tag: ## Tags git repository with latest version
	@{ \
		version=$$( echo ${VERSION} ) ; \
		set +e ; \
		git show-ref --quiet --verify "refs/tags/$$version" ; \
		res=$$? ; \
		set -e ; \
		if [[ ! $$res -eq 0 ]]; then \
			git tag -a $$version -m "git-tag $$version" ; \
		fi ; \
	}

.PHONY: tidy
tidy: ## Fetches dependencies
	@echo "Go Mod Vendor"
	$(GO) mod tidy
	$(GO) mod vendor
	@echo

.PHONY: echo
echo: ## Prints image name and version of the tool
	@echo "git-get ${VERSION} ${BUILD}"

.PHONY: release
release: ## Release application
	@{ \
		version=$$( echo ${VERSION} ) ; \
		exe=$$( echo ${EXE} ) ; \
		set +e ; \
		git show-ref --quiet --verify "refs/tags/$$version" ; \
		res=$$? ; \
		set -e ; \
		if [[ ! $$res -eq 0 ]]; then \
			git tag -a $$version -m "git-tag $$version" ; \
			git-chglog "$$version" > chglog.tmp ; \
			hub release create -F chglog.tmp "$$version" -a bin/$${exe}-$${version}-linux -a bin/$${exe}-$${version}-osx ; \
			rm -f chglog.tmp ; \
		fi ; \
	}
