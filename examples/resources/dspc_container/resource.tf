# Create a container
resource "dspc_container" "example" {
  name  = "my-container"
  image = "hello-world"
}

# Output the container details
output "container_id" {
  description = "The ID of the created container"
  value       = dspc_container.example.id
}

output "container_name" {
  description = "The name of the created container"
  value       = dspc_container.example.name
}
