# List all virtual machine groups
data "dspc_vm_groups" "all" {}

# Output all VMGroup names
output "vmg_names" {
  description = "List of all virtual machine group names"
  value       = [for vm in data.dspc_vm_groups.all.vm_groups : vm.name]
}

# Output all VMGroup URNs
output "vmg_urns" {
  description = "List of all virtual machine group URNs"
  value       = [for vmg in data.dspc_vm_groups.all.vm_groups : vmg.urn]
}

# Output all VMGroup statuses
output "vmg_statuses" {
  description = "Map of virtual machine group name to its current status"
  value       = { for vmg in data.dspc_vm_groups.all.vm_groups : vmg.name => vmg.status }
}

# Output count of VMGroups
output "vmg_count" {
  description = "Total number of virtual machine groups"
  value       = length(data.dspc_vm_groups.all.vm_groups)
}
