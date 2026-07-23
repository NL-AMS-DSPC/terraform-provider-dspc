# Get a container.
data "asc_container" "example" {
  name = "my-container"
}

# Output the container details
output "container_id" {
  description = "The ID of the created container"
  value       = data.asc_container.example.id
}

output "container_name" {
  description = "The name of the created container"
  value       = data.asc_container.example.name
}

output "container_tenant_id" {
  description = "The tenant that owns the container deployment"
  value       = data.asc_container.example.tenant_id
}
