terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}

provider "dspc" {
  # Configuration via environment variables (recommended for CI/CD)
  # DSPC_ENDPOINT="https://vm-deployer.example.com:8080"
  # DSPC_API_KEY="your-api-key-here"
  # DSPC_TIMEOUT="60"

  # Or configure directly (not recommended for production)
  # endpoint = "https://vm-deployer.example.com:8080"
  # api_key  = "your-api-key-here"
  # timeout  = 60
}
