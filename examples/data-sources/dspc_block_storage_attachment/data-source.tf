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