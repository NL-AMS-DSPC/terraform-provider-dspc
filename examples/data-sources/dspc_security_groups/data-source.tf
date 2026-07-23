# List all Security Groups
data "asc_security_groups" "all" {}

# Output all Security Group names
output "security_group_names" {
  description = "List of all Security Group names"
  value       = [for sg in data.asc_security_groups.all.security_groups : sg.name]
}

# Output count of Security Groups
output "security_group_count" {
  description = "Total number of Security Groups"
  value       = length(data.asc_security_groups.all.security_groups)
}
