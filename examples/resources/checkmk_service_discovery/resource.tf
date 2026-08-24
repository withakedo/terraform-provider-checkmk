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

resource "checkmk_service_discovery" "web" {
  host_name = checkmk_host.web.host_name
  mode      = "fix_all"

  depends_on = [checkmk_host.web]
}
