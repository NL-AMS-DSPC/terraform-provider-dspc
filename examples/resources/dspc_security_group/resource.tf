# Create a Security Group
resource "dspc_security_group" "example" {
  name = "restrict-backend-egress"
}

# Output the Security Group details
output "security_group_id" {
  description = "The ID of the created Security Group"
  value       = dspc_security_group.example.id
}

output "security_group_name" {
  description = "The name of the created Security Group"
  value       = dspc_security_group.example.name
}

# Delete a Security Group
#
# Option 1: Destroy only the security group resource:
#   terraform destroy -target=dspc_security_group.example
#
# Option 2: Remove the dspc_security_group resource block above from this file,
#   then run:
#   terraform apply
#
# Option 3: Destroy all resources managed by this configuration:
#   terraform destroy
#
# NOTE: Ensure all security rules and attachments are removed before
#   deleting the security group.
