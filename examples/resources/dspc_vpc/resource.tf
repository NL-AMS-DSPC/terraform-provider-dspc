# Create a VPC
resource "dspc_vpc" "example" {
  name = "my-vpc"

  tags = [
    {
      key   = "environment"
      value = "dev"
    }
  ]

  subnets = [
    {
      name = "my-vpc-public"
      cidr = "10.0.0.0/25"
      type = "public"
    }
  ]
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
