# List all available SKUs
data "dspc_skus" "all" {}

# Output all SKU names
output "sku_names" {
  description = "List of all available SKU names"
  value       = [for s in data.dspc_skus.all.skus : s.name]
}

# Output SKU cores keyed by name
output "sku_cores" {
  description = "Map of SKU name to its number of cores"
  value       = { for s in data.dspc_skus.all.skus : s.name => s.cores }
}

# Output count of SKUs
output "sku_count" {
  description = "Total number of available SKUs"
  value       = length(data.dspc_skus.all.skus)
}
