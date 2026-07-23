# Create a Security Group
resource "asc_security_group" "example" {
  name = "my-sg"
}

# Attach the Security Group to a Virtual Machine
resource "asc_security_group_attachment" "example" {
  security_group_name = asc_security_group.example.name
  target_type         = "VirtualMachine"
  target_name         = "my-vm"
}

# Output the attachment details
output "attachment_id" {
  description = "The ID of the security group attachment"
  value       = asc_security_group_attachment.example.id
}

output "attachment_name" {
  description = "The Kubernetes attachment resource name"
  value       = asc_security_group_attachment.example.attachment_name
}
