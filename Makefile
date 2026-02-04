# Makefile for DSPC Terraform Provider

.PHONY: build install test clean docs lint build-all tools help

# Linting
CONFIG_DIR := .git-lint-config
CONFIG_REPO := NL-AMS-DSPC/guidelines
CONFIG_BRANCH := main

.PHONY: build install test clean docs fmt lint

build: ## Build the provider binary
	go build -o terraform-provider-dspc

install: build ## Install the provider locally
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/NL-AMS-DSPC/dspc/1.0.0/linux_amd64/
	cp terraform-provider-dspc ~/.terraform.d/plugins/registry.terraform.io/NL-AMS-DSPC/dspc/1.0.0/linux_amd64/

# Run tests
test: ## Run tests
	TF_ACC=1 go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f terraform-provider-dspc
	rm -f coverage.out coverage.html

# Generate documentation
docs: ## Generate Docs
	go generate ./...

# Generate documentation only
docs-only: ## Generate Docs only
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name dspc

################ Linting and Testing ################

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Run linter
fetch-lint-config:
	@echo "Fetching centralized lint config via git (using your existing SSH credentials)..."
	@if [ ! -d $(CONFIG_DIR) ]; then \
		git clone --no-checkout --depth 1 --filter=blob:none --sparse \
			git@github.com:$(CONFIG_REPO).git $(CONFIG_DIR); \
	fi
	@cd $(CONFIG_DIR) && \
		git sparse-checkout set .golangci.yml .revive.toml && \
		git checkout $(CONFIG_BRANCH) -- .golangci.yml .revive.toml
	@cp $(CONFIG_DIR)/.golangci.yml $(CONFIG_DIR)/.revive.toml .
	@echo "Config files updated."


lint: fetch-lint-config ## Run golangci-lint
	@if [ ! -d $(CONFIG_DIR) ]; then \
		git clone --no-checkout --depth 1 --filter=blob:none --sparse \
			git@github.com:$(CONFIG_REPO).git $(CONFIG_DIR); \
	fi


	@echo "Installing/updating golangci-lint v2..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo "Running golangci-lint..."
	-@golangci-lint run ./...
	@echo "Linting complete"
	@rm -rf $(CONFIG_DIR) .golangci.yml .revive.toml

build-all: ## Build provider for multiple platforms
	GOOS=linux GOARCH=amd64 go build -o terraform-provider-dspc_linux_amd64
	GOOS=windows GOARCH=amd64 go build -o terraform-provider-dspc_windows_amd64.exe
	GOOS=darwin GOARCH=amd64 go build -o terraform-provider-dspc_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o terraform-provider-dspc_darwin_arm64

tools: ## Manage Go module dependencies
	go mod tidy
	go mod download
