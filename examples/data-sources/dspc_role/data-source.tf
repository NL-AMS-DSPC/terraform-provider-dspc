# Look up an existing role by name
data "dspc_role" "example" {
  name = "vm-operator"
}

# Output the role details
output "role_name" {
  description = "The name of the role"
  value       = data.dspc_role.example.name
}

output "role_actions" {
  description = "The permission actions assigned to the role"
  value       = data.dspc_role.example.actions
}
