# List all subnets for a given VPC
data "dspc_subnets" "all" {
  vpc_name = "my-vpc"
}

# Output all subnet names
output "subnet_names" {
  description = "List of all subnet names in the VPC"
  value       = [for subnet in data.dspc_subnets.all.subnets : subnet.name]
}

# Output all subnet IDs
output "subnet_ids" {
  description = "List of all subnet IDs in the VPC"
  value       = [for subnet in data.dspc_subnets.all.subnets : subnet.id]
}

# Output count of subnets
output "subnet_count" {
  description = "Total number of subnets in the VPC"
  value       = length(data.dspc_subnets.all.subnets)
}
