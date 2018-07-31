build:
	@CGO_ENABLED=0 GOOS=linux go build -o terraform-provider-bless

test:
	@TF_ACC=yes go test -cover -v ./...

packr:
	packr

release: packr
	./release

.PHONY: build test packr release
