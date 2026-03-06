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

# Look up an existing group by name
data "dspc_group" "example" {
  name = "platform-team"
}

# Output the group details
output "group_name" {
  description = "The name of the group"
  value       = data.dspc_group.example.name
}
