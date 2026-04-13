# Read an existing PostgreSQL instance
data "dspc_postgresql" "existing" {
  name = "my-postgres-db"
}

# Output the instance details
output "postgresql_size" {
  description = "The storage size of the PostgreSQL instance"
  value       = data.dspc_postgresql.existing.size
}

output "postgresql_version" {
  description = "The version of the PostgreSQL instance"
  value       = data.dspc_postgresql.existing.version
}

output "postgresql_vpc" {
  description = "The VPC of the PostgreSQL instance"
  value       = data.dspc_postgresql.existing.vpc
}
