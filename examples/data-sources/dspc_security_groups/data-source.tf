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
  # DSPC_API_KEY="your-api-key-here"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30
}

# List all Security Groups
data "dspc_security_groups" "all" {}

# Output all Security Group names
output "security_group_names" {
  description = "List of all Security Group names"
  value       = [for sg in data.dspc_security_groups.all.security_groups : sg.name]
}

# Output count of Security Groups
output "security_group_count" {
  description = "Total number of Security Groups"
  value       = length(data.dspc_security_groups.all.security_groups)
}
