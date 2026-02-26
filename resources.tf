#resource "dspc_virtual_machine" "example" {
#  name = "terraform-vm4"
#  sku_id = "gp-2"
#}

#resource "dspc_vpc" "example" {
#  name = "my-vpc-abc"
#  cidr = "10.5.0.0/24"
#}

# Create a Security Group
resource "dspc_security_group" "example" {
  name = "piotr-security-group"
}

# Output the Security Group details
output "sg_id" {
  description = "The ID of the created Security Group"
  value       = dspc_security_group.example.id
}

output "sg_name" {
  description = "The name of the created Security Group"
  value       = dspc_security_group.example.name
}

# First, create a Security Group
#resource "dspc_security_group" "example" {
#  name = "piotr-security-group2"
#}

# Add an ingress rule allowing TCP port 80 from a specific CIDR
resource "dspc_security_rule" "allow_http" {
  security_group_name = dspc_security_group.example.name
  direction           = "ingress"

  peers {
    ip_block_cidr = "10.0.0.0/24"
  }

  ports {
    protocol = "TCP"
    port     = 8000
  }
}

# Add an egress rule allowing TCP port 443
resource "dspc_security_rule" "allow_https_out" {
  security_group_name = dspc_security_group.example.name
  direction           = "egress"

  ports {
    protocol = "TCP"
    port     = 4430
  }
}

# Output the rule details
output "http_rule_id" {
  description = "The ID of the HTTP ingress rule"
  value       = dspc_security_rule.allow_http.id
}

#data "dspc_security_groups" "all" {}

#output "piotr_sg" {
#  value = [for sg in data.dspc_security_groups.all.security_groups : sg if sg.name == "piotr-security-group"]
#}
