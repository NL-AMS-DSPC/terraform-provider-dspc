data "dspc_file_storage_access" "example" {
  file_storage_name = "my-file-storage"
  target_type       = "VirtualMachine"
  target_name       = "my-vm"
}

output "file_storage_access_id" {
  description = "The ID of the file storage access entry."
  value       = data.dspc_file_storage_access.example.id
}
