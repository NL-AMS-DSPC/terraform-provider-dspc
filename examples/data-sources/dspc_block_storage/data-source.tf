data "asc_block_storage" "example" {
  name="my-example-block"
}

# Output the block details
output "block_id" {
  description = "The ID of the requested block"
  value       = data.asc_block_storage.example.id

}
output "block_name" {
  description = "The ID of the requested block"
  value       = data.asc_block_storage.example.name
}

output "block_size" {
  description = "The size of the requested block"
  value       = data.asc_block_storage.example.size
}