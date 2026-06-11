---
page_title: "dspc_mssql Data Source - dspc"
subcategory: "Managed Databases"
description: |-
  Retrieves a Microsoft SQL Server instance from the DSPC platform.
---

# dspc_mssql (Data Source)

Retrieves a Microsoft SQL Server instance from the DSPC platform.

## Example Usage

```terraform
data "dspc_mssql" "existing" {
  name = "my-mssql-db"
}

output "mssql_sku_size" {
  description = "The SKU size of the MSSQL instance"
  value       = data.dspc_mssql.existing.sku_size
}

output "mssql_version" {
  description = "The version of the MSSQL instance"
  value       = data.dspc_mssql.existing.version
}

output "mssql_vpc_id" {
  description = "The VPC ID of the MSSQL instance"
  value       = data.dspc_mssql.existing.vpc_id
}
```

## Schema

### Required

- `name` (String) Unique name of the database instance to retrieve.

### Read-Only

- `sku_size` (String) SKU size per node instance, e.g. `gp-2`, `gp-4`, etc.
- `version` (String) Version of the database engine. One of: `MSSQL_2025_17`, `MSSQL_2022_16`, `MSSQL_2019_15`, `MSSQL_2017_14`.
- `vpc_id` (String) GUID of the VPC network where this database is deployed.
- `tags` (List of Object) Tags applied to the database instance. Each tag object contains:
  - `key` (String) Tag key.
  - `value` (String) Tag value.

