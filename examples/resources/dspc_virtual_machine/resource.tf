terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}

provider "dspc" {
  # Configure via environment variables:
  # DSPC_ENDPOINT="https://vm-deployer.example.com:8080"
  # DSPC_API_KEY="your-api-key-here"
}

# Create a virtual machine
resource "dspc_virtual_machine" "example" {
  name = "my-example-vm"
}

# Output the VM details
output "vm_id" {
  description = "The ID of the created virtual machine"
  value       = dspc_virtual_machine.example.id
}

output "vm_name" {
  description = "The name of the created virtual machine"
  value       = dspc_virtual_machine.example.name
}
