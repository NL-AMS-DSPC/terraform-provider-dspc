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
  # DSPC_ENDPOINT="https://network-orchestrator.example.com:8080"
  # DSPC_NAMESPACE="corp-namespace"
  # DSPC_USERNAME="your-client-id"
  # DSPC_PASSWORD="your-client-secret"
  # DSPC_AUTH_URL="https://auth.example.com"
  # DSPC_ORG="your-org"
}

# Create a Security Group
resource "dspc_security_group" "example" {
  name = "my-security-group"
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
