# Create a subnet within a VPC
resource "asc_subnet" "example" {
  name     = "my-subnet"
  vpc_name = asc_vpc.example.name
  vpc_id = asc_vpc.example.id
  cidr     = "10.0.0.0/25"
  type     = "public"

  tags = {
    env  = "demo"
    team = "platform"
  }
}

# Output the subnet details
output "subnet_id" {
  description = "The ID of the created subnet"
  value       = asc_subnet.example.id
}

output "subnet_name" {
  description = "The name of the created subnet"
  value       = asc_subnet.example.name
}

output "subnet_vpc_name" {
  description = "The VPC name the subnet belongs to"
  value       = asc_subnet.example.vpc_name
}

output "subnet_cidr" {
  description = "The CIDR block of the created subnet"
  value       = asc_subnet.example.cidr
}

output "subnet_type" {
  description = "The type of the created subnet"
  value       = asc_subnet.example.type
}

# Delete a subnet from a VPC
#
# Option 1: Destroy only the subnet resource:
#   terraform destroy -target=asc_subnet.example
#
# Option 2: Remove the asc_subnet resource block above from this file,
#   then run:
#   terraform apply
#
# Option 3: Destroy all resources managed by this configuration:
#   terraform destroy
#
# NOTE: The parent VPC (asc_vpc.example) will remain intact
#   when deleting only the subnet.
