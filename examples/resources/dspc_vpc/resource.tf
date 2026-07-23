# Create a VPC
resource "asc_vpc" "example" {
  name = "my-vpc"
  cidr = "10.0.0.0/24"
}

# Output the VPC details
output "vpc_id" {
  description = "The ID of the created VPC"
  value       = asc_vpc.example.id
}

output "vpc_name" {
  description = "The name of the created VPC"
  value       = asc_vpc.example.name
}

output "vpc_cidr" {
  description = "The CIDR block of the created VPC"
  value       = asc_vpc.example.cidr
}

# Delete a VPC
#
# Option 1: Remove all subnets first, then destroy only the VPC resource:
#   terraform destroy -target=asc_vpc.example
#
# Option 2: Remove the asc_vpc resource block above from this file,
#   then run:
#   terraform apply
#
# Option 3: Destroy all resources managed by this configuration:
#   terraform destroy
#
# NOTE: A VPC cannot be deleted while it still has subnets.
#   Ensure all subnets are removed before deleting the VPC.
