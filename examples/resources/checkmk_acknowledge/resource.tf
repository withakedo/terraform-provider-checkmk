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

# Acknowledge a problem on the whole host
resource "checkmk_acknowledge" "web_down" {
  host_name = checkmk_host.web.host_name
  comment   = "Investigating, ticket OPS-1234"
}
