terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}

provider "dspc" {
  # REQUIRED: Configure via environment variables (recommended)
  # DSPC_ENDPOINT="https://network-orchestrator.example.com:8080"
  # DSPC_NAMESPACE="corp-namespace"
  # DSPC_API_KEY="your-api-key-here"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30
}

# List rules for a specific Security Group
data "dspc_security_rules" "example" {
  security_group_name = "db-access"
}

# Output ingress rules
output "ingress_rules" {
  description = "Ingress rules for the Security Group"
  value       = data.dspc_security_rules.example.ingress
}

# Output egress rules
output "egress_rules" {
  description = "Egress rules for the Security Group"
  value       = data.dspc_security_rules.example.egress
}

# Output count of all rules
output "total_rule_count" {
  description = "Total number of rules"
  value       = length(data.dspc_security_rules.example.ingress) + length(data.dspc_security_rules.example.egress)
}
