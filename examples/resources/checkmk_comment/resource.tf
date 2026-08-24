terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"
}

resource "checkmk_host" "web" {
  host_name = "web-01.example.com"
  folder    = "/"

  attributes = {
    ipaddress = "10.1.2.3"
  }
}

# Comment on the whole host
resource "checkmk_comment" "note" {
  host_name = checkmk_host.web.host_name
  comment   = "Migrated to new datacenter rack B12 on 2026-08-24"
}
