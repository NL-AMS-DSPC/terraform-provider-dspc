# Provision a managed OpenShift cluster via ASC cluster-service.
resource "asc_cluster" "example" {
  name    = "demo-cluster"
  domain  = "example.com"
  version = "4.16.5"
  image   = "rhcos-4.16.5"

  control_plane = {
    replicas = 3
    sku_id   = "cp-medium"
  }

  workers = {
    replicas = 3
    sku_id   = "worker-medium"
  }

  vpc = {
    name = "demo-vpc"
    subnets = {
      pods     = { name = "demo-pods" }
      services = { name = "demo-services" }
    }
  }

  tags = {
    env  = "demo"
    team = "platform"
  }

  # Write-only: read from a secret store, never committed.
  pull_secret = var.pull_secret
  ssh_key     = var.ssh_key
}

variable "pull_secret" {
  description = "Red Hat pull secret JSON used to render the install-config."
  type        = string
  sensitive   = true
}

variable "ssh_key" {
  description = "SSH public key authorized on cluster nodes."
  type        = string
  sensitive   = true
}

output "cluster_urn" {
  description = "URN assigned by cluster-service."
  value       = asc_cluster.example.urn
}

output "cluster_status" {
  description = "Lifecycle status of the cluster."
  value       = asc_cluster.example.status
}
