provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_manifest" "function" {
  manifest = {
    apiVersion = "serving.knative.dev/v1"
    kind       = "Service"
    metadata = {
      name      = "stan-hello-tf"
      namespace = "development"
    }
    spec = {
      traffic = [
        {
          percent        = 100
          latestRevision = true
        }
      ]
      template = {
        metadata = {
          annotations = {
            "autoscaling.knative.dev/minScale" = "1"            # Min/max replicas
            "autoscaling.knative.dev/maxScale" = "2"
            "serving.knative.dev/rolloutDuration" = "120s"      # Rolling update strategies
            "serving.knative.dev/rolloutTarget" = "10"          # Percent traffic to new revision initially
          }
        }
        spec = {
          containers = [
            {
              image = "gcr.io/knative-samples/helloworld-go"    # Container image (registry URL)
              env = [                                           # Environment variables
                { name = "TARGET", value = "Terraform" }
              ]
              ports = [
                { containerPort = 8080 }
              ]
              resources = {                                     # CPU & RAM limits
                limits = {
                  cpu    = "500m"       # 0.5 CPU cores
                  memory = "512Mi"      # 512 MiB of memory
                }
                requests = {
                  cpu    = "100m"       # 0.1 CPU cores
                  memory = "128Mi"      # 128 MiB of memory
                }
              }
            }
          ]
        }
      }
    }
  }
}