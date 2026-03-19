terraform {
  required_providers {
    dspc = {
      source = "NL-AMS-DSPC/dspc"
    }
  }
}
 


# Create a function with basic configuration
resource "dspc_function" "example" {
  name   = "my-example-function"
  sku_id = "gp-2"
}

# Output the function details
output "function_id" {
  description = "The ID of the created function"
  value       = dspc_function.example.id
}

output "function_name" {
  description = "The name of the created function"
  value       = dspc_function.example.name
}

output "function_status" {
  description = "The current status of the function"
  value       = dspc_function.example.status
}