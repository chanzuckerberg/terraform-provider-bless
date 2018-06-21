build:
	@go build -o terraform-provider-bless

test:
	@TF_ACC=yes go test -v ./...
