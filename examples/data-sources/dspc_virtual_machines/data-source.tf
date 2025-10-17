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
  # DSPC_API_KEY="your-api-key-here"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30
}

# List all virtual machines
data "dspc_virtual_machines" "all" {}

# Output all VM names
output "vm_names" {
  description = "List of all virtual machine names"
  value       = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.name]
}

# Output all VM IDs
output "vm_ids" {
  description = "List of all virtual machine IDs"
  value       = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.id]
}

# Output count of VMs
output "vm_count" {
  description = "Total number of virtual machines"
  value       = length(data.dspc_virtual_machines.all.virtual_machines)
}
