# List all virtual machines
data "dspc_virtual_machines" "all" {}

# Output all VM names
output "vm_names" {
  description = "List of all virtual machine names"
  value       = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.name]
}

# Output all VM URNs
output "vm_urns" {
  description = "List of all virtual machine URNs"
  value       = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.urn]
}

# Output all VM statuses
output "vm_statuses" {
  description = "Map of virtual machine name to its current status"
  value       = { for vm in data.dspc_virtual_machines.all.virtual_machines : vm.name => vm.status }
}

# Output count of VMs
output "vm_count" {
  description = "Total number of virtual machines"
  value       = length(data.dspc_virtual_machines.all.virtual_machines)
}
