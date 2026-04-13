# Create a PostgreSQL instance
resource "dspc_postgresql" "example" {
  name    = "my-postgres-db"
  size    = "1Gi"
  version = "POSTGRES_17"
  vpc     = "my-vpc"
}

# Create a PostgreSQL instance with tags
resource "dspc_postgresql" "tagged" {
  name    = "my-tagged-postgres"
  size    = "2Gi"
  version = "POSTGRES_16"
  vpc     = "my-vpc"

  tags = [
    {
      key   = "env"
      value = "production"
    },
    {
      key   = "team"
      value = "platform"
    },
  ]
}

# Output the instance details
output "postgresql_name" {
  description = "The name of the created PostgreSQL instance"
  value       = dspc_postgresql.example.name
}

output "postgresql_version" {
  description = "The version of the created PostgreSQL instance"
  value       = dspc_postgresql.example.version
}
