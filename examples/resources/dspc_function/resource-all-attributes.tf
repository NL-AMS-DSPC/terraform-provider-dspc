# Create a function with all available configuration attributes
resource "dspc_function" "example" {
  name  = "dspc-function-all-attributes-tf"
  image = "gcr.io/knative-samples/helloworld-go"

  # Container port configuration
  port = 8080

  # Concurrency settings
  concurrency {
    limit = 10
  }

  # Environment variables
  env = [
    {
      name  = "TARGET"
      value = "World"
    },
    {
      name  = "ENV"
      value = "production"
    }
  ]

  # Health check configuration
  health_checks {
    liveness {
      failure_threshold     = 3
      initial_delay_seconds = 15
      path                  = "/health"
      period_seconds        = 20
      port                  = 8080
      timeout_seconds       = 5
    }

    readiness {
      failure_threshold     = 3
      initial_delay_seconds = 10
      path                  = "/ready"
      period_seconds        = 10
      port                  = 8080
      timeout_seconds       = 3
    }
  }

  # Resource limits and requests
  resources {
    cpu_limit      = "500m"
    cpu_request    = "100m"
    memory_limit   = "512Mi"
    memory_request = "128Mi"
  }

  # Pull the image from a private registry (write-only: never returned on read).
  # server is optional — derived from the image when omitted. Changing these forces recreation.
  registry_auth = {
    server   = "harbor.example.com"
    username = "puller"
    password = var.registry_password
  }

  # Runtime secrets exposed as env vars; values are stored in OpenBao.
  # Write-only: never returned on read. Changing them forces recreation.
  secrets = [
    {
      env_name = "DATABASE_PASSWORD"
      value    = var.database_password
    },
    {
      env_name = "API_KEY"
      value    = var.api_key
    }
  ]

  # Tags/Labels
  tags = [
    {
      key   = "env"
      value = "local"
    },
    {
      key   = "team"
      value = "platform"
    },
    {
      key   = "cost-center"
      value = "engineering"
    }
  ]
}

variable "registry_password" {
  type      = string
  sensitive = true
}

variable "database_password" {
  type      = string
  sensitive = true
}

variable "api_key" {
  type      = string
  sensitive = true
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

output "function_url" {
  description = "The URL of the function"
  value       = dspc_function.example.url
}

output "latest_ready_revision" {
  description = "The latest ready revision of the function"
  value       = dspc_function.example.latest_ready_revision
}

output "created_at" {
  description = "The creation timestamp of the function"
  value       = dspc_function.example.created_at
}

output "updated_at" {
  description = "The last update timestamp of the function"
  value       = dspc_function.example.updated_at
}