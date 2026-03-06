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

# Create a group to add a member to
resource "dspc_group" "example" {
  name = "platform-team"
}

# Add a user to the group
resource "dspc_group_member" "example" {
  group_name = dspc_group.example.name
  user_id    = "user-uuid-1234"
}

# Output the membership details
output "membership_id" {
  description = "The ID of the group membership (group_name:user_id)"
  value       = dspc_group_member.example.id
}
