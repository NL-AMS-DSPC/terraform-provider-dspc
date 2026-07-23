terraform {
  required_providers {
    asc = {
      source  = "asc/asc"
      version = "~> 1.0"
    }
  }
}

provider "asc" {
  # REQUIRED: Configure via environment variables (recommended for CI/CD)
  # ASC_ENDPOINT="https://api.example.com"
  # ASC_NAMESPACE="corp-namespace"
  # ASC_USERNAME="auth-service-client-id"
  # ASC_PASSWORD="auth-service-client-secret"
  # ASC_AUTH_URL="https://auth-service.example.com"
  # ASC_ORG="organization-realm"
  # ASC_TIMEOUT="60"  # Optional, defaults to 30

  # OR configure directly (not recommended for production)
  # endpoint  = "https://api.example.com"               # REQUIRED
  # username  = "auth-service-client-id"                # REQUIRED (Auth service client ID)
  # password  = "auth-service-client-secret"            # REQUIRED (Auth service client secret)
  # auth_url  = "https://auth-service.example.com"      # REQUIRED (Auth service base URL)
  # org       = "organization-realm"                    # REQUIRED (Auth service realm)
  # namespace = "corp-namespace"                        # REQUIRED
  # timeout   = 60                                      # Optional, defaults to 30
}
