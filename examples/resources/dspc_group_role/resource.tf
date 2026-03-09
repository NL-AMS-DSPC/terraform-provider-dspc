# Create a role and a group, then assign the role to the group
resource "dspc_role" "example" {
  name = "vm-operator"
  actions = [
    "vm:CreateVM",
    "vm:DeleteVM",
    "vm:ListVMs",
  ]
}

resource "dspc_group" "example" {
  name = "platform-team"
}

resource "dspc_group_role" "example" {
  group_name = dspc_group.example.name
  role_name  = dspc_role.example.name
}

# Output the assignment details
output "group_role_id" {
  description = "The ID of the group role assignment"
  value       = dspc_group_role.example.id
}
