# Create a virtual machine
resource "asc_block_storage_attachment" "example" {
  vm_name = "my-example-vm"
  block_storage_name = "my-example-block"
}

# Output the block attachment details
output "block_attachment_id" {
  description = "The ID of the requested block attachment"
  value       = asc_block_storage_attachment.example.id
}

output "block_attachment_name" {
  description = "The ID of the requested block attachment"
  value       = asc_block_storage_attachment.example.block_storage_name
}

output "block_attachment_vm_name" {
  description = "The virtual machine name of the requested block attachment"
  value       = asc_block_storage_attachment.example.vm_name
}