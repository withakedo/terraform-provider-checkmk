# Multiple host groups for infrastructure organization
resource "checkmk_host_group" "databases" {
  name  = "databases"
  alias = "Database Servers"
}

resource "checkmk_host_group" "loadbalancers" {
  name  = "loadbalancers"
  alias = "Load Balancers"
}

resource "checkmk_host_group" "network_devices" {
  name  = "network_devices"
  alias = "Network Infrastructure"
}

resource "checkmk_host_group" "storage" {
  name  = "storage"
  alias = "Storage Systems"
}

resource "checkmk_host_group" "kubernetes" {
  name  = "kubernetes"
  alias = "Kubernetes Nodes"
}
