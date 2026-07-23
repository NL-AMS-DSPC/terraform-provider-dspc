# Create a Security Group first
resource "asc_security_group" "example" {
  name = "db-access"
}

# Add an egress rule allowing TCP traffic to a CIDR block
resource "asc_security_rule" "allow_db_egress" {
  security_group_name = asc_security_group.example.name
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
resource "asc_security_rule" "allow_app_ingress" {
  security_group_name = asc_security_group.example.name
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
  value       = asc_security_rule.allow_db_egress.id
}

output "ingress_rule_id" {
  description = "The ID of the ingress rule"
  value       = asc_security_rule.allow_app_ingress.id
}
