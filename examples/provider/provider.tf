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
  # DSPC_ENDPOINT="https://api.example.com"
  # DSPC_NAMESPACE="corp-namespace"
  # DSPC_USERNAME="auth-service-client-id"
  # DSPC_PASSWORD="auth-service-client-secret"
  # DSPC_AUTH_URL="https://auth-service.example.com"
  # DSPC_ORG="organization-realm"
  # DSPC_TIMEOUT="60"  # Optional, defaults to 30

  # OR configure directly (not recommended for production)
  # endpoint  = "https://api.example.com"               # REQUIRED
  # username  = "auth-service-client-id"                # REQUIRED (Auth service client ID)
  # password  = "auth-service-client-secret"            # REQUIRED (Auth service client secret)
  # auth_url  = "https://auth-service.example.com"      # REQUIRED (Auth service base URL)
  # org       = "organization-realm"                    # REQUIRED (Auth service realm)
  # namespace = "corp-namespace"                        # REQUIRED
  # timeout   = 60                                      # Optional, defaults to 30
}
