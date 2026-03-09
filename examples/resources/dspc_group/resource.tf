# Create an authorization group
resource "dspc_group" "example" {
  name = "platform-team"
}

# Output the group details
output "group_name" {
  description = "The name of the created group"
  value       = dspc_group.example.name
}
