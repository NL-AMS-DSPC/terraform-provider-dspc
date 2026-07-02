# Create a container
resource "dspc_container" "example" {
  name  = "my-container"
  image = "hello-world"
  port  = 8080

  # Pull from a private registry (write-only: never returned on read)
  registry_auth = {
    server   = "harbor.example.com"
    username = "puller"
    password = var.registry_password
  }

  # Runtime secrets exposed as env vars (write-only)
  secrets = [
    {
      env_name = "DB_PASSWORD"
      value    = var.db_password
    },
    {
      env_name = "API_TOKEN"
      value    = var.api_token
    },
  ]
}

variable "registry_password" {
  type      = string
  sensitive = true
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "api_token" {
  type      = string
  sensitive = true
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

output "container_tenant_id" {
  description = "The tenant that owns the container deployment"
  value       = dspc_container.example.tenant_id
}
