export GOFLAGS=-mod=vendor
export GO111MODULE=on

setup: ## setup development dependencies
	./.godownloader-packr.sh -d v1.24.1
	curl -sfL https://raw.githubusercontent.com/chanzuckerberg/bff/master/download.sh | sh
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -sfL https://raw.githubusercontent.com/reviewdog/reviewdog/master/install.sh| sh
.PHONY: setup

lint: ## run the fast go linters on the diff from master
	./bin/reviewdog -conf .reviewdog.yml  -diff "git diff master"
.PHONY: lint

lint-ci: ## run the linteres and annotate a PR (useful only for running in CI)
	./bin/reviewdog -conf .reviewdog.yml  -reporter=github-pr-review
.PHONY: lint-ci

lint-all: ## run all the linters
	# doesn't seem to be a way to get reviewdog to not filter by diff
	./bin/golangci-lint run
.PHONY: lint-all

build: packr
	@CGO_ENABLED=0 GOOS=linux go build -o terraform-provider-bless
.PHONY:  build

test: deps packr
	@TF_ACC=yes go test -cover -v ./...
.PHONY: test

test-ci: packr
	@TF_ACC=yes go test -cover -v ./...
.PHONY: test-ci

deps:
	go mod tidy
	go mod vendor
.PHONY: deps

packr: clean
	packr
.PHONY: packr

clean: ## clean the repo
	rm terraform-provider-bless 2>/dev/null || true
	go clean
	rm -rf dist 2>/dev/null || true
	packr clean
	rm coverage.out 2>/dev/null || true
.PHONY: clean

release: ## run a release
	bff bump
	git push
	goreleaser release --rm-dist
.PHONY: release