terraform {
  required_providers {
    dspc = {
      source  = "dspc/dspc"
      version = "~> 1.0"
    }
  }
}

provider "dspc" {
  # REQUIRED: Configure via environment variables (recommended for CI/CD)
  # DSPC_ENDPOINT="https://vm-deployer.example.com:8080"
  # DSPC_API_KEY="your-api-key-here"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30

  # OR configure directly (not recommended for production)
  # endpoint = "https://vm-deployer.example.com:8080"  # REQUIRED
  # api_key  = "your-api-key-here"                     # REQUIRED
  # timeout  = 60                                      # Optional, defaults to 30
}
