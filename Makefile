# Makefile for DSPC Terraform Provider

.PHONY: build install test clean docs fmt lint

# Build the provider
build:
	go build -o terraform-provider-dspc

# Install the provider locally
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/dspc/dspc/1.0.0/linux_amd64/
	cp terraform-provider-dspc ~/.terraform.d/plugins/registry.terraform.io/dspc/dspc/1.0.0/linux_amd64/

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f terraform-provider-dspc
	rm -f coverage.out coverage.html

# Generate documentation
docs:
	go generate ./...

# Generate documentation only
docs-only:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name dspc

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run all checks
check: fmt lint test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o terraform-provider-dspc_linux_amd64
	GOOS=windows GOARCH=amd64 go build -o terraform-provider-dspc_windows_amd64.exe
	GOOS=darwin GOARCH=amd64 go build -o terraform-provider-dspc_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o terraform-provider-dspc_darwin_arm64

# Development setup
dev-setup:
	go mod tidy
	go mod download
