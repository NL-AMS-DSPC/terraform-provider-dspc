# Quick Start Guide

This guide will help you get started with the DSPC Terraform Provider quickly.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- Access to a running DSPC VM Deployer API (default: `http://localhost:8080`)

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/dspc/dpsc-terraform-provider.git
cd dpsc-terraform-provider

# Build the provider
make build

# Install locally
make install
```

### Option 2: Use Terraform Registry (Future)

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

## Basic Usage

1. **Create a Terraform configuration file** (`main.tf`):

```hcl
terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}

provider "dspc" {
  endpoint = "http://localhost:8080"
}

resource "dspc_virtual_machine" "example" {
  name = "my-first-vm"
}

output "vm_id" {
  value = dspc_virtual_machine.example.id
}
```

2. **Initialize Terraform**:

```bash
terraform init
```

3. **Plan the deployment**:

```bash
terraform plan
```

4. **Apply the configuration**:

```bash
terraform apply
```

5. **Verify the VM was created**:

```bash
# List all VMs
data "dspc_virtual_machines" "all" {}

output "all_vms" {
  value = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.name]
}
```

## Common Commands

```bash
# Initialize Terraform
terraform init

# Plan changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy

# Show current state
terraform show

# List resources
terraform state list
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `endpoint` | `http://localhost:8080` | DSPC VM Deployer API endpoint |
| `timeout` | `30` | API timeout in seconds |

## Troubleshooting

### Connection Issues

If you get connection errors:

1. Verify the DSPC VM Deployer is running
2. Check the endpoint URL
3. Ensure network connectivity

### API Errors

Check the API logs and ensure:
- The API is accessible
- Authentication is configured (if required)
- The platform is supported

### Debug Mode

Enable debug logging:

```bash
export TF_LOG=DEBUG
terraform plan
```

## Next Steps

- Explore the [full documentation](README.md)

## Support

- Contact the DSPC team
- Check the internal documentation
