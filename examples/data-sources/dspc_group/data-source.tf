# Look up an existing group by name
data "asc_group" "example" {
  name = "platform-team"
}

# Output the group details
output "group_name" {
  description = "The name of the group"
  value       = data.asc_group.example.name
}
