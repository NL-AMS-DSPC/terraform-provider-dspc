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

# Create a role with specific permissions
resource "dspc_role" "example" {
  name = "vm-operator"
  actions = [
    "vm:CreateVM",
    "vm:DeleteVM",
    "vm:ListVMs",
  ]
}

# Output the role details
output "role_name" {
  description = "The name of the created role"
  value       = dspc_role.example.name
}

output "role_actions" {
  description = "The permission actions assigned to the role"
  value       = dspc_role.example.actions
}
