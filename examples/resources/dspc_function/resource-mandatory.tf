# Create a function with basic configuration
resource "asc_function" "example" {
  name  = "asc-function-mandatory-tf"
  image = "gcr.io/knative-samples/helloworld-go"
}

# Output the function details
output "function_id" {
  description = "The ID of the created function"
  value       = asc_function.example.id
}

output "function_name" {
  description = "The name of the created function"
  value       = asc_function.example.name
}

output "function_status" {
  description = "The current status of the function"
  value       = asc_function.example.status
}

output "function_url" {
  description = "The URL of the created function"
  value       = asc_function.example.url
}