provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_manifest" "function" {
  manifest = {
    apiVersion = "serving.knative.dev/v1"
    kind       = "Service"
    metadata = {
      name      = "dspc-functions-demo-app"
      namespace = "development"
    }
    spec = {
      template = {
        spec = {
          containers = [
            {
              image = "gcr.io/knative-samples/helloworld-go"    # Container image (registry URL)
              env = [
                { name = "TARGET", value = "Terraform" }
              ]
            }
          ]
        }
      }
    }
  }
}