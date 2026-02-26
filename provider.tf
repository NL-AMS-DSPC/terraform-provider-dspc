terraform {
  required_providers {
    dspc = {
      source  = "nl-ams-dspc/dspc"
      version = "~> 1.0.0"
    }
  }
}

provider "dspc" {
  endpoint  = "https://api.apps.omfcd2bf61f633a630.uksouth.aroapp.io"  # REQUIRED
  auth_url = "https://keycloak-ingress-keycloak-mgmt.apps.omgnucglm0dd36e236.uksouth.aroapp.io"
  username = "[terraform]-something"
  password = "aPnCw64NUKn0XIDPB27kp8Yp6MtA9dyW"
  timeout   = 60                                      # Optional, defaults to 30
  org = "dspc-demo"
  namespace = "project-piotr"
}
