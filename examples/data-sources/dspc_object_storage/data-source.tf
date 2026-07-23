data "asc_object_storage" "example" {
  name = "my-example-object-storage"
}

# Output the object_storage details
output "object_storage_id" {
  description = "The ID of the requested object-storage"
  value       = data.asc_object_storage.example.id
}

output "object_storage_name" {
  description = "The name of the requested object-storage"
  value       = data.asc_object_storage.example.name
}

output "object_storage_max_size" {
  description = "The size of the requested object-storage"
  value       = data.asc_object_storage.example.max_size
}
