build: packr
	@CGO_ENABLED=0 GOOS=linux go build -o terraform-provider-bless

test: packr
	@TF_ACC=yes go test -cover -v ./...

packr:
	packr

release: packr ## run a release
	bff bump
	git push
	goreleaser release --rm-dist

.PHONY: build test packr release
