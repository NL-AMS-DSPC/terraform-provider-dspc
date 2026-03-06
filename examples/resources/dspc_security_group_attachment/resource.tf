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
  # DSPC_ENDPOINT="https://api.example.com"
  # DSPC_NAMESPACE="corp-namespace"
  # DSPC_USERNAME="auth-service-client-id"
  # DSPC_PASSWORD="auth-service-client-secret"
  # DSPC_AUTH_URL="https://auth-service.example.com"
  # DSPC_ORG="organization-realm"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30
}

# Create a Security Group
resource "dspc_security_group" "example" {
  name = "my-sg"
}

# Attach the Security Group to a Virtual Machine
resource "dspc_security_group_attachment" "example" {
  security_group_name = dspc_security_group.example.name
  target_type         = "VirtualMachine"
  target_name         = "my-vm"
}

# Output the attachment details
output "attachment_id" {
  description = "The ID of the security group attachment"
  value       = dspc_security_group_attachment.example.id
}

output "attachment_name" {
  description = "The Kubernetes attachment resource name"
  value       = dspc_security_group_attachment.example.attachment_name
}
