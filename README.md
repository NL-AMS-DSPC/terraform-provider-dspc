# DSPC Terraform Provider

A Terraform provider for managing virtual machines via the DSPC VM Deployer API.

## Features

- **VM Management**: Create, read, and delete virtual machines
- **Authentication**: API key support with Bearer token authentication
- **Environment Variables**: Configure via environment variables for CI/CD
- **Multi-platform**: Supports Linux, Windows, and macOS (amd64/arm64)
- **Terraform Registry**: Ready for publishing to Terraform Registry

## Quick Start

### Installation

#### From Terraform Registry (Future)

```hcl
terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}
```

#### Manual Installation

1. Download the binary for your platform from [releases](../../releases)
2. Place it in your Terraform plugins directory:
   - **Windows**: `%APPDATA%\terraform.d\plugins\registry.terraform.io\dspc\dspc\1.0.0\windows_amd64\`
   - **macOS**: `~/.terraform.d/plugins/registry.terraform.io/dspc/dspc/1.0.0/darwin_amd64/`
   - **Linux**: `~/.terraform.d/plugins/registry.terraform.io/dspc/dspc/1.0.0/linux_amd64/`
3. Rename to `terraform-provider-dspc` (or `terraform-provider-dspc.exe` on Windows)

### Configuration

```hcl
provider "dspc" {
  endpoint = "http://localhost:8080"  # Default endpoint
  timeout  = 60
  api_key  = "your-api-key-here"  # Optional, can use DSPC_API_KEY env var
}
```

### Environment Variables

```bash
export DSPC_ENDPOINT="https://vm-deployer.example.com:8080"
export DSPC_TIMEOUT="60"
export DSPC_API_KEY="your-api-key-here"
```

### Basic Usage

```hcl
# Create a VM
resource "dspc_virtual_machine" "example" {
  name = "my-first-vm"
}

# List all VMs
data "dspc_virtual_machines" "all" {}

output "vm_names" {
  value = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.name]
}
```

## Development

### Prerequisites

- Go 1.21+
- Terraform 1.0+
- Access to DSPC VM Deployer API

### Building

```bash
# Build the provider
make build

# Install locally for testing
make install

# Run tests
make test

# Generate documentation
make docs
```

### Testing

```bash
# Run unit tests
go test ./...

# Run tests with coverage
make test-coverage
```

## API Compatibility

This provider currently supports the minimal DSPC VM API:

- **Create VM**: `POST /virtualmachine` with `{"vmName": "..."}`
- **Delete VM**: `DELETE /virtualmachine` with `{"vmName": "..."}`
- **List VMs**: `GET /virtualmachine`

### Authentication

The provider sends `Authorization: Bearer <token>` headers with all requests. The current DSPC API doesn't validate these tokens yet, but the provider is ready for when authentication is implemented.

## Versioning

This provider follows [Semantic Versioning](https://semver.org/):

- **v1.x.x**: Minimal VM API support (name field only)
- **v2.x.x**: Extended VM API support (cpu, memory, disk, etc.)
- **v3.x.x**: Additional resource types (containers, storage, etc.)

## Publishing to Terraform Registry

### Prerequisites

1. GitHub repository connected to Terraform Registry
2. GPG key configured for signing releases
3. GitHub secrets configured:
   - `GPG_FINGERPRINT`
   - `GPG_PRIVATE_KEY`
   - `GPG_PASSPHRASE`

### Release Process

1. Create and push a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Sign the release with GPG
   - Terraform Registry will pull the release

### Manual Registry Setup

1. Go to [registry.terraform.io](https://registry.terraform.io)
2. Sign in with GitHub
3. Click "Publish a Provider"
4. Select your repository
5. Configure webhook settings
6. Registry will automatically publish on new releases

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to ensure all checks pass
6. Submit a pull request

## License

This project is licensed under the Mozilla Public License Version 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- Documentation: [docs/](docs/)
- Issues: [GitHub Issues](../../issues)
- Changelog: [CHANGELOG.md](CHANGELOG.md)
