# DSPC Terraform Provider

A Terraform provider for managing virtual machines and block storage via the DSPC API.

## Features

- **VM Management**: Create, read, update, and delete virtual machines
- **Network Management**: Create and manage VPCs and subnets
- **Block Storage Management**: Create, read, update, and delete block storage volumes
- **Block Storage Attachments**: Attach/detach block storage to virtual machines
- **Authentication**: OAuth2/JWT authentication via auth service
- **Environment Variables**: Configure via environment variables for CI/CD
- **Flexible Service Paths**: Configurable API service path prefixes for different environments
- **Multi-platform**: Supports Linux, Windows, and macOS (amd64/arm64)
- **Terraform Registry**: Published at `registry.terraform.io/NL-AMS-DSPC/dspc`

## Quick Start

### Installation

#### From Terraform Registry

```hcl
terraform {
  required_providers {
    dspc = {
      source  = "NL-AMS-DSPC/dspc"
      version = "~> 1.0"
    }
  }
}
```

#### Manual Installation

1. Download the binary for your platform from [releases](../../releases)
2. Place it in your Terraform plugins directory:
   - **Windows**: `%APPDATA%\terraform.d\plugins\registry.terraform.io\NL-AMS-DSPC\dspc\1.0.0\windows_amd64\`
   - **macOS**: `~/.terraform.d/plugins/registry.terraform.io/NL-AMS-DSPC/dspc/1.0.0/darwin_amd64/`
   - **Linux**: `~/.terraform.d/plugins/registry.terraform.io/NL-AMS-DSPC/dspc/1.0.0/linux_amd64/`
3. Rename to `terraform-provider-dspc` (or `terraform-provider-dspc.exe` on Windows)

### Configuration

```hcl
provider "dspc" {
  endpoint  = "https://api.example.com"                # REQUIRED - API endpoint
  auth_url  = "https://auth-service.example.com"       # REQUIRED - Auth service URL
  org       = "organization-realm"                     # REQUIRED - Auth service realm
  username  = "auth-service-client-id"                 # REQUIRED - Auth service client ID
  password  = "auth-service-client-secret"             # REQUIRED - Auth service client secret
  namespace = "my-namespace"                           # REQUIRED - Resource namespace
  timeout   = 60                                       # OPTIONAL - Request timeout in seconds (default: 30)
}
```

### Environment Variables

#### Core Configuration

```bash
export DSPC_ENDPOINT="https://api.example.com"
export DSPC_AUTH_URL="https://auth-service.example.com"
export DSPC_ORG="organization-realm"
export DSPC_USERNAME="auth-service-client-id"
export DSPC_PASSWORD="auth-service-client-secret"
export DSPC_NAMESPACE="my-namespace"
export DSPC_TIMEOUT="60"  # Optional
```

#### Advanced: Service Path Configuration

The provider supports customizing API service path prefixes for different deployments or API versions. This is useful when working with environments behind Envoy gateway or custom API routing:

```bash
# Override default service paths (optional)
export DSPC_VM_PATH_PREFIX="/api/vm"              # Default: /api/vm
export DSPC_NETWORK_PATH_PREFIX="/api/network"    # Default: /api/network
export DSPC_STORAGE_PATH_PREFIX="/api/vm"         # Default: /api/vm
```

**Use cases:**
- API versioning: `DSPC_VM_PATH_PREFIX="/v2/virtualmachines"`
- Custom routing: `DSPC_NETWORK_PATH_PREFIX="/custom/network"`
- Different environments: Production vs staging API paths

### Basic Usage

#### Virtual Machines

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

#### Block Storage

```hcl
# Create block storage
resource "dspc_block_storage" "data" {
  name = "my-data-volume"
  size = "10Gi"
}

# Get block storage details
data "dspc_block_storage" "existing" {
  name = "my-data-volume"
}

output "block_size" {
  value = data.dspc_block_storage.existing.size
}
```

#### Block Storage Attachments

```hcl
# Attach block storage to VM
resource "dspc_block_storage_attachment" "attach" {
  vm_name            = dspc_virtual_machine.example.name
  block_storage_name = dspc_block_storage.data.name
}

# Query attachment
data "dspc_block_storage_attachment" "check" {
  vm_name            = "my-first-vm"
  block_storage_name = "my-data-volume"
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

This provider supports the DSPC API with the following default endpoints. Service path prefixes can be customized via environment variables (see [Advanced: Service Path Configuration](#advanced-service-path-configuration)).

### Virtual Machines
- **Create VM**: `POST /api/vm/v1/namespaces/{namespace}/virtualmachines`
- **Get VM**: `GET /api/vm/v1/namespaces/{namespace}/virtualmachines/{name}`
- **List VMs**: `GET /api/vm/v1/namespaces/{namespace}/virtualmachines`
- **Delete VM**: `DELETE /api/vm/v1/namespaces/{namespace}/virtualmachines/{name}`

### Block Storage
- **Create Block**: `POST /api/vm/v1/namespaces/{namespace}/blocks`
- **Get Block**: `GET /api/vm/v1/namespaces/{namespace}/blocks/{name}`
- **List Blocks**: `GET /api/vm/v1/namespaces/{namespace}/blocks`
- **Update Block**: `PUT /api/vm/v1/namespaces/{namespace}/blocks/{name}`
- **Delete Block**: `DELETE /api/vm/v1/namespaces/{namespace}/blocks/{name}`

### Block Storage Attachments
- **Attach Block**: `POST /api/vm/v1/namespaces/{namespace}/blocks/{block}/attach/{vm}`
- **List Attachments**: `GET /api/vm/v1/namespaces/{namespace}/virtualmachines/{vm}/blocks`
- **Detach Block**: `DELETE /api/vm/v1/namespaces/{namespace}/blocks/{block}/attach/{vm}`

### Network (VPC & Subnets)
- **Create VPC**: `POST /api/network/v1/namespaces/{namespace}/vpcs`
- **Get VPC**: `GET /api/network/v1/namespaces/{namespace}/vpcs/{name}`
- **List VPCs**: `GET /api/network/v1/namespaces/{namespace}/vpcs`
- **Delete VPC**: `DELETE /api/network/v1/namespaces/{namespace}/vpcs/{name}`
- **Create Subnet**: `POST /api/network/v1/namespaces/{namespace}/vpcs/{vpc}/subnets`
- **List Subnets**: `GET /api/network/v1/namespaces/{namespace}/vpcs/{vpc}/subnets`
- **Delete Subnet**: `DELETE /api/network/v1/namespaces/{namespace}/vpcs/{vpc}/subnets/{subnet}`

### Functions
- **Create Function**: `POST /api/functions/v1/namespaces/{namespace}/functions`
- **Get Function**: `GET /api/functions/v1/namespaces/{namespace}/functions/{name}`
- **List Function**: `GET /api/functions/v1/namespaces/{namespace}/functions`

### Authentication

The provider authenticates using OAuth2 client credentials flow:
1. Obtains JWT token from auth service: `POST {auth_url}/realms/{org}/protocol/openid-connect/token`
2. Caches token with 30-second expiration buffer
3. Automatically refreshes token when expired
4. Sends `Authorization: Bearer <jwt-token>` with all API requests

## Versioning

This provider follows [Semantic Versioning](https://semver.org/):

- **v1.x.x**: Current version - VM and block storage management with OAuth2 authentication
- **v2.x.x**: Future - Extended VM configuration (cpu, memory, SKUs, etc.)
- **v3.x.x**: Future - Additional resource types (networks, load balancers, etc.)

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

