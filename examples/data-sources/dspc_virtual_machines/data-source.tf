# List all virtual machines
data "dspc_virtual_machines" "all" {}

# Output all VM names
output "vm_names" {
  description = "List of all virtual machine names"
  value       = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.name]
}

# Output all VM IDs
output "vm_ids" {
  description = "List of all virtual machine IDs"
  value       = [for vm in data.dspc_virtual_machines.all.virtual_machines : vm.id]
}

# Output count of VMs
output "vm_count" {
  description = "Total number of virtual machines"
  value       = length(data.dspc_virtual_machines.all.virtual_machines)
}
