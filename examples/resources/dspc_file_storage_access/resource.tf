resource "dspc_file_storage" "example" {
  name = "my-file-storage"
  size = "100Gi"
}

resource "dspc_virtual_machine" "example" {
  name   = "my-vm"
  sku_id = "standard-2cpu-4gb"
}

resource "dspc_file_storage_access" "example" {
  file_storage_name = dspc_file_storage.example.name
  target_type       = "VirtualMachine"
  target_name       = dspc_virtual_machine.example.name
}
