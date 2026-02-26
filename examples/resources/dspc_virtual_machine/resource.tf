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

# Create a virtual machine with autoscaling
resource "dspc_virtual_machine" "example" {
  name   = "my-example-vm"
  sku_id = "medium"

  # Optional: Configure autoscaling
  autoscaling {
    min_replicas                         = 1
    max_replicas                         = 5
    target_cpu_utilization_percentage    = 70
    target_memory_utilization_percentage = 80

    # Optional: Enable scale-to-zero
    # enable_scale_to_zero = true
    # idle_replicas        = 1
    # scale_to_zero_after  = 300
  }
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

output "vm_status" {
  description = "The current status of the virtual machine"
  value       = dspc_virtual_machine.example.status
}

output "vm_replicas" {
  description = "The current number of VM replicas"
  value       = dspc_virtual_machine.example.replicas
}

