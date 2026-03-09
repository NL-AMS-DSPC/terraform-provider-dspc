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

