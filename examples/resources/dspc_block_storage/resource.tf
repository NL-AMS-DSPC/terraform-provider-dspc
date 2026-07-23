# Create a virtual machine
resource "asc_block_storage" "example" {
  name = "my-example-block"
  size = "5Gi"
}

# Output the block details
output "block_id" {
  description = "The ID of the created block"
  value       = asc_block_storage.example.id
}

output "block_name" {
  description = "The name of the created block"
  value       = asc_block_storage.example.name
}

output "block_size" {
  description = "The size of the created block"
  value       = asc_block_storage.example.size
}