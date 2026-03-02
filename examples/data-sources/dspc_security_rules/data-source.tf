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
  # DSPC_USERNAME="your-client-id"
  # DSPC_PASSWORD="your-client-secret"
  # DSPC_AUTH_URL="https://auth.example.com"
  # DSPC_ORG="your-org"
}

# List all rules for a security group
data "dspc_security_rules" "example" {
  security_group_name = "my-security-group"
}

# Output the ingress rules
output "ingress_rules" {
  description = "All ingress rules for the security group"
  value       = data.dspc_security_rules.example.ingress_rules
}

# Output the egress rules
output "egress_rules" {
  description = "All egress rules for the security group"
  value       = data.dspc_security_rules.example.egress_rules
}
