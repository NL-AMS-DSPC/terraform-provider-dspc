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

# Create a VPC
resource "dspc_vpc" "example" {
  name = "my-vpc"
  cidr = "10.0.0.0/24"
}

# Output the VPC details
output "vpc_id" {
  description = "The ID of the created VPC"
  value       = dspc_vpc.example.id
}

output "vpc_name" {
  description = "The name of the created VPC"
  value       = dspc_vpc.example.name
}

output "vpc_cidr" {
  description = "The CIDR block of the created VPC"
  value       = dspc_vpc.example.cidr
}

# Delete a VPC
#
# Option 1: Remove all subnets first, then destroy only the VPC resource:
#   terraform destroy -target=dspc_vpc.example
#
# Option 2: Remove the dspc_vpc resource block above from this file,
#   then run:
#   terraform apply
#
# Option 3: Destroy all resources managed by this configuration:
#   terraform destroy
#
# NOTE: A VPC cannot be deleted while it still has subnets.
#   Ensure all subnets are removed before deleting the VPC.
