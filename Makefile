SHELL := /bin/bash
GO 		:= GO15VENDOREXPERIMENT=1 GO111MODULE=on GOPROXY=https://proxy.golang.org go

VERSION:="0.0.1"
BUILD:=`git rev-parse --short HEAD`
TIME:=`date`
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: clean mod fmt test build

.PHONY: build
## build: builds the binary file
build:
	@go build -ldflags="-X 'github.com/isindir/git-get/version.Version=v${VERSION}' \
		-X 'github.com/isindir/git-get/version.Commit=${BUILD}' \
		-X 'github.com/isindir/git-get/version.Time=${TIME}' " \
		-o ./bin/git-get main.go

.PHONY: run
# run: runs main help
run:
	go run main.go

.PHONY: test
## test: placeholder to run unit tests
test:
	@echo "Running unit tests"
	@mkdir -p bin
	go test -cover -coverprofile=bin/c.out ./...
	go tool cover -html=bin/c.out -o bin/coverage.html
	@echo

.PHONY: check
## check: runs linting
check:
	@echo "Linting"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@echo

.PHONY: fmt
## fmt: formats go code
fmt:
	@echo "Formatting"
	@gofmt -l -w $(SRC)
	@echo

.PHONY: clean
## clean: removes build artifacts from source code
clean:
	@echo "Cleaning"
	@rm -fr bin
	@rm -fr vendor
	@rm -fr chglog.tmp
	@echo

.PHONY: repo-tag
## repo-tag: tags git repository with latest version
repo-tag:
	@git tag -a ${VERSION} -m "git-tag ${VERSION}"

.PHONY: mod
## mod: fetches dependencies
mod:
	@echo "Go Mod Vendor"
	$(GO) mod tidy
	# $(GO) mod vendor
	@echo

.PHONY: echo
## echo: prints image name and version of the operator
echo:
	@echo "git-get ${VERSION} ${BUILD}"

.PHONY: release
## release: release application
release: repo-tag
	@git-chglog "${VERSION}" > chglog.tmp
	@hub release create -F chglog.tmp "${VERSION}" -a bin/git-get
	@rm -f chglog.tmp

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: moq
moq:
	$(MAKE) -C ./gitget moq
