resource "dspc_file_storage" "example" {
  name = "my-file-storage"
  size = "100Gi"
}

output "file_storage_id" {
  description = "The ID of the created file storage."
  value       = dspc_file_storage.example.id
}

output "file_storage_nfs_mount_path" {
  description = "The NFS mount path for the file storage."
  value       = dspc_file_storage.example.nfs_mount_path
}
