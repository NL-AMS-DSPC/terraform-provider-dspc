terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}

provider "dspc" {
  # REQUIRED: Configure via environment variables (recommended)
  # DSPC_ENDPOINT="https://vm-deployer.example.com:8080"
  # DSPC_NAMESPACE="corp-namespace"
  # DSPC_API_KEY="your-api-key-here"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30
}

# Create a virtual machine
resource "dspc_block_storage" "example" {
  name = "my-example-block"
  size = "5Gi"
}

# Output the block details
output "block_id" {
  description = "The ID of the created block"
  value       = dspc_block_storage.example.id
}

output "block_name" {
  description = "The name of the created block"
  value       = dspc_block_storage.example.name
}

output "block_size" {
  description = "The size of the created block"
  value       = dspc_block_storage.example.size
}