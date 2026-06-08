---
page_title: "dspc_mssql Resource - dspc"
subcategory: "Managed Databases"
description: |-
  Manages a Microsoft SQL Server instance in the DSPC platform.
---

# dspc_mssql (Resource)

Manages a Microsoft SQL Server instance in the DSPC platform.

## Example Usage

```terraform
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
```

## Schema

### Required

- `name` (String) Unique name for the database instance. Must be 1-63 lowercase alphanumeric characters or hyphens, and must start and end with an alphanumeric character.
- `sku_size` (String) SKU size per instance node, e.g. `gp-2`, `gp-4`, etc.
- `version` (String) Version of the database engine. One of: `MSSQL_2025_17`, `MSSQL_2022_16`, `MSSQL_2019_15`, `MSSQL_2017_14`.
- `vpc_id` (String) GUID of the VPC network where this database should be deployed.

### Optional

- `additional_configuration` (Block, Optional) Additional configuration options for the MSSQL instance.
  - `license_key` (String, Sensitive) License key for the MSSQL instance.
- `tags` (List of Object, Optional) Tags to apply to the database instance. Each tag object supports:
  - `key` (String, Required) Tag key. Must be a qualified name (max 316 chars total).
  - `value` (String, Required) Tag value (max 63 chars). May be empty.

## Import

Import an existing MSSQL instance by its name:

```shell
terraform import dspc_mssql.example my-mssql-db
```

