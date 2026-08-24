terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

# Provider with strict ETag validation
provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"

  activate = "auto"

  # Enforce strict ETag validation
  # Activation will fail if resources were modified outside Terraform
  force_foreign_changes = false
}

resource "checkmk_host" "critical" {
  host_name = "critical-server"
  folder    = "/"

  attributes = {
    alias     = "Critical Infrastructure"
    ipaddress = "10.0.1.100"
  }
}

# Use strict ETag validation when:
# - You want to detect external modifications
# - Multiple teams/tools manage CheckMK
# - Compliance requires change tracking
