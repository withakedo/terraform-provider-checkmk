terraform {
  required_providers {
    checkmk = {
      source = "withake-it/checkmk"
    }
  }
}

# Provider with manual activation (default behavior)
provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"

  # activate = "manual" is the default
  # Changes must be manually activated in CheckMK
}

resource "checkmk_host" "staging" {
  host_name = "staging-host"
  folder    = "/staging"

  attributes = {
    alias     = "Staging Server"
    ipaddress = "192.168.2.10"
    site      = "production"
  }
}

resource "checkmk_host" "production" {
  host_name = "prod-host"
  folder    = "/production"

  attributes = {
    alias     = "Production Server"
    ipaddress = "192.168.1.10"
    site      = "production"
  }
}

# With manual activation, you can batch create multiple hosts
# and activate all changes at once in the CheckMK UI
