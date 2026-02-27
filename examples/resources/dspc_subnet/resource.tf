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

# Create a subnet within a VPC
resource "dspc_subnet" "example" {
  name     = "my-subnet"
  vpc_name = dspc_vpc.example.name
  cidr     = "10.0.0.0/25"
  type     = "public"
}

# Output the subnet details
output "subnet_id" {
  description = "The ID of the created subnet"
  value       = dspc_subnet.example.id
}

output "subnet_name" {
  description = "The name of the created subnet"
  value       = dspc_subnet.example.name
}

output "subnet_vpc_name" {
  description = "The VPC name the subnet belongs to"
  value       = dspc_subnet.example.vpc_name
}

output "subnet_cidr" {
  description = "The CIDR block of the created subnet"
  value       = dspc_subnet.example.cidr
}

output "subnet_type" {
  description = "The type of the created subnet"
  value       = dspc_subnet.example.type
}

# Delete a subnet from a VPC
#
# Option 1: Destroy only the subnet resource:
#   terraform destroy -target=dspc_subnet.example
#
# Option 2: Remove the dspc_subnet resource block above from this file,
#   then run:
#   terraform apply
#
# Option 3: Destroy all resources managed by this configuration:
#   terraform destroy
#
# NOTE: The parent VPC (dspc_vpc.example) will remain intact
#   when deleting only the subnet.
