terraform {
  required_providers {
    checkmk = {
      source = "blackmesaltd/checkmk"
    }
  }
}

provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"
}

resource "checkmk_host" "example" {
  host_name = "example-host"
  folder    = "/"

  attributes = {
    alias     = "Example Host"
    ipaddress = "192.168.1.100"
  }
}
