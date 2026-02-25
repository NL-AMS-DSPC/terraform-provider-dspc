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

# List all VPCs
data "dspc_vpcs" "all" {}

# Output all VPC names
output "vpc_names" {
  description = "List of all VPC names"
  value       = [for vpc in data.dspc_vpcs.all.vpcs : vpc.name]
}

# Output all VPC IDs
output "vpc_ids" {
  description = "List of all VPC IDs"
  value       = [for vpc in data.dspc_vpcs.all.vpcs : vpc.id]
}

# Output count of VPCs
output "vpc_count" {
  description = "Total number of VPCs"
  value       = length(data.dspc_vpcs.all.vpcs)
}
