# Create a load balancer
resource "asc_load_balancer" "example" {
  name      = "my-load-balancer"
  scheme    = "some-scheme"
  sku_id    = "lb-small"
  vpc_id    = asc_vpc.example.id
  subnet_id = asc_subnet.example.id
  algorithm = "round-robin"

  listeners = [
    {
      protocol = "TCP"
      port     = 80
    }
  ]

  backends = [
    {
      ip   = "10.0.0.10"
      port = 8080
    },
    {
      ip   = "10.0.0.11"
      port = 8080
    }
  ]

  health_check = {
    protocol            = "TCP"
    port                = 8080
    interval_seconds    = 10
    timeout_seconds     = 5
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  tags = {
    environment = "production"
  }
}

output "load_balancer_id" {
  description = "The ID of the created load balancer"
  value       = asc_load_balancer.example.id
}