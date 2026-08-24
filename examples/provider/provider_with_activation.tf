terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

# Configure the CheckMK Provider with automatic activation
provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"

  # Activation mode: "auto", "manual" (default)
  # - "auto": Automatically activate changes after create/update/delete
  # - "manual": Changes are staged but require manual activation in CheckMK UI
  activate = "auto"

  # Force foreign changes: bypass ETag validation (default: true)
  # When true, uses If-Match: * to allow activating changes made outside Terraform
  # When false, enforces strict ETag validation (may fail if resources modified externally)
  force_foreign_changes = true

  # Wait time after activation in seconds (default: 5)
  # Increase if experiencing eventual consistency issues
  activation_wait_time = 5
}

resource "checkmk_host" "example" {
  host_name = "auto-activated-host"
  folder    = "/"

  attributes = {
    alias     = "Auto-Activated Host"
    ipaddress = "192.168.1.100"
  }
}
