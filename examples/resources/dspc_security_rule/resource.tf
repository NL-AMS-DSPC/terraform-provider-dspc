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

# Create a Security Group first
resource "dspc_security_group" "example" {
  name = "db-access"
}

# Add an egress rule allowing TCP traffic to a CIDR block
resource "dspc_security_rule" "allow_db_egress" {
  security_group_name = dspc_security_group.example.name
  direction           = "egress"

  ports {
    protocol = "TCP"
    port     = 5432
  }

  peers {
    ip_block {
      cidr   = "10.0.0.0/24"
      except = ["10.0.0.1/32"]
    }
  }
}

# Add an ingress rule allowing traffic from specific pods
resource "dspc_security_rule" "allow_app_ingress" {
  security_group_name = dspc_security_group.example.name
  direction           = "ingress"

  ports {
    protocol = "TCP"
    port     = 8080
  }

  peers {
    pod_selector = {
      app = "frontend"
    }
  }
}

# Output the rule details
output "egress_rule_id" {
  description = "The ID of the egress rule"
  value       = dspc_security_rule.allow_db_egress.id
}

output "ingress_rule_id" {
  description = "The ID of the ingress rule"
  value       = dspc_security_rule.allow_app_ingress.id
}
