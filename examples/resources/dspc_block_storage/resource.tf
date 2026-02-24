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
  # DSPC_ENDPOINT="https://api.example.com"
  # DSPC_NAMESPACE="corp-namespace"
  # DSPC_USERNAME="auth-service-client-id"
  # DSPC_PASSWORD="auth-service-client-secret"
  # DSPC_AUTH_URL="https://auth-service.example.com"
  # DSPC_ORG="organization-realm"
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