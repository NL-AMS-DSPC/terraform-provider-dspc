provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_manifest" "function" {
  manifest = {
    apiVersion = "serving.knative.dev/v1"
    kind       = "Service"
    metadata = {
      name      = "stan-hello-tf-mandatory"
      namespace = "development"
    }
    spec = {
      template = {
        spec = {
          containers = [
            {
              image = "gcr.io/knative-samples/helloworld-go"    # Container image (registry URL)
            }
          ]
        }
      }
    }
  }
}