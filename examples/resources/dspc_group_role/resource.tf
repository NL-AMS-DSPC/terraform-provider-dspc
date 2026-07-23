# Create a role and a group, then assign the role to the group
resource "asc_role" "example" {
  name = "vm-operator"
  actions = [
    "vm:CreateVM",
    "vm:DeleteVM",
    "vm:ListVMs",
  ]
}

resource "asc_group" "example" {
  name = "platform-team"
}

resource "asc_group_role" "example" {
  group_name = asc_group.example.name
  role_name  = asc_role.example.name
}

# Output the assignment details
output "group_role_id" {
  description = "The ID of the group role assignment"
  value       = asc_group_role.example.id
}
