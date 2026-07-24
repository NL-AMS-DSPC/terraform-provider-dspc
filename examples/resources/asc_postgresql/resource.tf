# Create a PostgreSQL instance
resource "asc_postgresql" "example" {
  name    = "my-postgres-db"
  sku_size    = "gp-2"
  version = "POSTGRES_17"
  vpc_id     = "00000000-0000-0000-0000-000000000000"
}

# Create a PostgreSQL instance with tags
resource "asc_postgresql" "tagged" {
  name    = "my-tagged-postgres"
  sku_size    = "gp-4"
  version = "POSTGRES_16"
  vpc_id     = "00000000-0000-0000-0000-000000000000"

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
  value       = asc_postgresql.example.name
}

output "postgresql_version" {
  description = "The version of the created PostgreSQL instance"
  value       = asc_postgresql.example.version
}
