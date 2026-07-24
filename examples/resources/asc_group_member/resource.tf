# Create a group to add a member to
resource "asc_group" "example" {
  name = "platform-team"
}

# Add a user to the group
resource "asc_group_member" "example" {
  group_name = asc_group.example.name
  user_id    = "user-uuid-1234"
}

# Output the membership details
output "membership_id" {
  description = "The ID of the group membership (group_name:user_id)"
  value       = asc_group_member.example.id
}
