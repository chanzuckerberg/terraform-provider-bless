export GOFLAGS=-mod=vendor
export GO111MODULE=on

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