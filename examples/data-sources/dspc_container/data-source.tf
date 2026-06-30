# Get a container.
data "dspc_container" "example" {
  name = "my-container"
}

# Output the container details
output "container_id" {
  description = "The ID of the created container"
  value       = data.dspc_container.example.id
}

output "container_name" {
  description = "The name of the created container"
  value       = data.dspc_container.example.name
}

output "container_tenant_id" {
  description = "The tenant that owns the container deployment"
  value       = data.dspc_container.example.tenant_id
}
