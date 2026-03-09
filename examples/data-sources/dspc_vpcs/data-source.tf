# List all VPCs
data "dspc_vpcs" "all" {}

# Output all VPC names
output "vpc_names" {
  description = "List of all VPC names"
  value       = [for vpc in data.dspc_vpcs.all.vpcs : vpc.name]
}

# Output all VPC IDs
output "vpc_ids" {
  description = "List of all VPC IDs"
  value       = [for vpc in data.dspc_vpcs.all.vpcs : vpc.id]
}

# Output count of VPCs
output "vpc_count" {
  description = "Total number of VPCs"
  value       = length(data.dspc_vpcs.all.vpcs)
}
