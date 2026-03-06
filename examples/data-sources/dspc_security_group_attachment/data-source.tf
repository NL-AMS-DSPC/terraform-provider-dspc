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
