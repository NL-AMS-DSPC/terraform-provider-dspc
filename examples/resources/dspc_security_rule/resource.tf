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

# First, create a Security Group
resource "dspc_security_group" "example" {
  name = "my-security-group"
}

# Add an ingress rule allowing TCP port 80 from a specific CIDR
resource "dspc_security_rule" "allow_http" {
  security_group_name = dspc_security_group.example.name
  direction           = "ingress"

  peers {
    ip_block_cidr = "10.0.0.0/24"
  }

  ports {
    protocol = "TCP"
    port     = 80
  }
}

# Add an egress rule allowing TCP port 443
resource "dspc_security_rule" "allow_https_out" {
  security_group_name = dspc_security_group.example.name
  direction           = "egress"

  ports {
    protocol = "TCP"
    port     = 443
  }
}

# Output the rule details
output "http_rule_id" {
  description = "The ID of the HTTP ingress rule"
  value       = dspc_security_rule.allow_http.id
}
