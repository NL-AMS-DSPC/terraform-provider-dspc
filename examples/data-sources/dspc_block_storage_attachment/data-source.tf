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

data "dspc_block_storage_attachment" "example" {
  block_storage_name="my-example-block"
  vm_name="my-example-vm"
}

# Output the block attachment details
output "block_attachment_id" {
  description = "The ID of the requested block attachment"
  value       = data.dspc_block_storage_attachment.example.id
}

output "block_attachment_name" {
  description = "The ID of the requested block attachment"
  value       = data.dspc_block_storage_attachment.example.block_storage_name
}

output "block_attachment_vm_name" {
  description = "The virtual machine name of the requested block attachment"
  value       = data.dspc_block_storage_attachment.example.vm_name
}