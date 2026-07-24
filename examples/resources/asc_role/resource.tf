# Create a role with specific permissions
resource "asc_role" "example" {
  name = "vm-operator"
  actions = [
    "vm:CreateVM",
    "vm:DeleteVM",
    "vm:ListVMs",
  ]
}

# Output the role details
output "role_name" {
  description = "The name of the created role"
  value       = asc_role.example.name
}

output "role_actions" {
  description = "The permission actions assigned to the role"
  value       = asc_role.example.actions
}
