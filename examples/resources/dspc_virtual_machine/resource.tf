# Create a virtual machine with autoscaling
resource "dspc_virtual_machine" "example" {
  name   = "my-example-vm"
  sku_id = "medium"
  vpc_id = "vpc-id"
  image = "vm-image"
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

