# Create a basic MSSQL instance
resource "dspc_mssql" "example" {
  name     = "my-mssql-db"
  sku_size = "gp-2"
  version  = "MSSQL_2022_16"
  vpc_id   = "00000000-0000-0000-0000-000000000000"
}

# Create an MSSQL instance with tags
resource "dspc_mssql" "tagged" {
  name     = "my-tagged-mssql"
  sku_size = "gp-4"
  version  = "MSSQL_2025_17"
  vpc_id   = "00000000-0000-0000-0000-000000000000"

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

# Create an MSSQL instance with a license key
resource "dspc_mssql" "licensed" {
  name     = "my-licensed-mssql"
  sku_size = "gp-8"
  version  = "MSSQL_2022_16"
  vpc_id   = "00000000-0000-0000-0000-000000000000"

  additional_configuration {
    license_key = var.mssql_license_key
  }
}

# Output the instance details
output "mssql_name" {
  description = "The name of the created MSSQL instance"
  value       = dspc_mssql.example.name
}

output "mssql_version" {
  description = "The version of the created MSSQL instance"
  value       = dspc_mssql.example.version
}

