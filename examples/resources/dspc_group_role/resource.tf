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

# Create a role and a group, then assign the role to the group
resource "dspc_role" "example" {
  name = "vm-operator"
  actions = [
    "vm:CreateVM",
    "vm:DeleteVM",
    "vm:ListVMs",
  ]
}

resource "dspc_group" "example" {
  name = "platform-team"
}

resource "dspc_group_role" "example" {
  group_name = dspc_group.example.name
  role_name  = dspc_role.example.name
}

# Output the assignment details
output "group_role_id" {
  description = "The ID of the group role assignment (group_name:role_name)"
  value       = dspc_group_role.example.id
}
