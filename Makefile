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
.PHONY: deps

packr: clean
	./bin/packr
.PHONY: packr

clean: ## clean the repo
	rm terraform-provider-bless 2>/dev/null || true
	go clean
	rm -rf dist 2>/dev/null || true
	./bin/packr clean
	rm coverage.out 2>/dev/null || true
.PHONY: clean

check-release-prereqs:
# ifndef KEYBASE_KEY_ID
# 	$(error KEYBASE_KEY_ID is undefined)
# endif
.PHONY: check-release-prereqs

release: check-release-prereqs ## run a release
	./bin/bff bump
	git push
	goreleaser release
.PHONY: release

release-prerelease: check-release-prereqs build ## release to github as a 'pre-release'
	version=`./$(BASE_BINARY_NAME) -version`; \
	git tag v"$$version"; \
	git push
	git push --tags
	goreleaser release -f .goreleaser.prerelease.yml --debug --rm-dist
.PHONY: release-prerelease