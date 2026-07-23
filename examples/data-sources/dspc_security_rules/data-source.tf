# List rules for a specific Security Group
data "asc_security_rules" "example" {
  security_group_name = "db-access"
}

# Output ingress rules
output "ingress_rules" {
  description = "Ingress rules for the Security Group"
  value       = data.asc_security_rules.example.ingress
}

# Output egress rules
output "egress_rules" {
  description = "Egress rules for the Security Group"
  value       = data.asc_security_rules.example.egress
}

# Output count of all rules
output "total_rule_count" {
  description = "Total number of rules"
  value       = length(data.asc_security_rules.example.ingress) + length(data.asc_security_rules.example.egress)
}
