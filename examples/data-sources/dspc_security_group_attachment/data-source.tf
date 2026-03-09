data "dspc_security_group_attachment" "example" {
  security_group_name = "my-sg"
  attachment_name     = "my-vm-my-sg-attach"
}

# Output the attachment details
output "attachment_id" {
  description = "The ID of the security group attachment"
  value       = data.dspc_security_group_attachment.example.id
}

output "attachment_sg_name" {
  description = "The security group name"
  value       = data.dspc_security_group_attachment.example.security_group_name
}
