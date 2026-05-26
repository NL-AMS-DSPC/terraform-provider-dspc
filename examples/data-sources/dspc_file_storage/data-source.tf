data "dspc_file_storage" "example" {
  name = "my-file-storage"
}

output "file_storage_size" {
  description = "The size of the file storage."
  value       = data.dspc_file_storage.example.size
}

output "file_storage_nfs_mount_path" {
  description = "The NFS mount path for the file storage."
  value       = data.dspc_file_storage.example.nfs_mount_path
}
