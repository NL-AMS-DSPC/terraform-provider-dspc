# Create an object storage
resource "dspc_object_sorage" "example" {
  name = "my-object-storage"
  quota = {
    max_size = "1Gi"
  }
}

# Output the object storage details
output "object_storage_name" {
  description = "The name of the created object storage"
  value       = dspc_object_storage.example.name
}

output "object_storage_size" {
  description = "The size of the created object storage"
  value       = dspc_object_storage.example.quota.max_size
}
