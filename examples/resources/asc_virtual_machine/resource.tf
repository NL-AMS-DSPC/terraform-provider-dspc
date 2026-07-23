# Create a virtual machine
resource "asc_virtual_machine" "example" {
  name   = "my-example-vm"
  sku_id = "medium"
  vpc_id = "vpc-id"
  image  = "vm-image"

  tags = {
    environment = "production"
    team        = "platform"
  }

  customization = {
    cloud_init = {
      user_data = <<-EOT
        #cloud-config
        hostname: my-example-vm
      EOT
    }
  }

  enable_logging = true
}

# Output the VM details
output "vm_urn" {
  description = "The URN of the created virtual machine"
  value       = asc_virtual_machine.example.urn
}

output "vm_name" {
  description = "The name of the created virtual machine"
  value       = asc_virtual_machine.example.name
}

output "vm_status" {
  description = "The current status of the virtual machine"
  value       = asc_virtual_machine.example.status
}

output "vm_attached_volumes" {
  description = "The volumes attached to the virtual machine"
  value       = asc_virtual_machine.example.attached_volumes
}