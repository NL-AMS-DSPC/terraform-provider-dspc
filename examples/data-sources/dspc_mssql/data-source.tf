# Read an existing MSSQL instance
data "asc_mssql" "existing" {
  name = "my-mssql-db"
}

# Output the instance details
output "mssql_sku_size" {
  description = "The SKU size of the MSSQL instance"
  value       = data.asc_mssql.existing.sku_size
}

output "mssql_version" {
  description = "The version of the MSSQL instance"
  value       = data.asc_mssql.existing.version
}

output "mssql_vpc_id" {
  description = "The VPC ID of the MSSQL instance"
  value       = data.asc_mssql.existing.vpc_id
}

