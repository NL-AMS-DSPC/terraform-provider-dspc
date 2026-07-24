# Create a virtual machine group
resource "asc_vm_group" "example" {
  name   = "my-example-vm-group"
  sku_id = "medium"
  vpc_id = "vpc-id"
  image  = "vm-image"

  tags = {
    environment = "production"
    team        = "platform"
  }

  autoscaling_policy = {
    min_replicas = 1
    max_replicas = 5

    cron = {
      timezone         = "UTC"
      start            = "0 8 * * *"
      end              = "0 18 * * *"
      desired_replicas = 3
    }
  }

  customization = {
    cloud_init = {
      user_data = <<-EOT
        #cloud-config
        hostname: my-example-vm-group
      EOT
    }
  }

}

# Output the VMGroup details
output "vmg_urn" {
  description = "The URN of the created virtual machine group"
  value       = asc_vm_group.example.urn
}

output "vmg_name" {
  description = "The name of the created virtual machine group"
  value       = asc_vm_group.example.name
}

output "vmg_status" {
  description = "The current status of the virtual machine group"
  value       = asc_vm_group.example.status
}

output "vmg_attached_volumes" {
  description = "The volumes attached to the virtual machine group"
  value       = asc_vm_group.example.attached_volumes
}