terraform {
  required_providers {
    checkmk = {
      source = "withake-it/checkmk"
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

# Host downtime
resource "checkmk_downtime" "maintenance" {
  host_name  = checkmk_host.web.host_name
  start_time = "2026-09-01T22:00:00Z"
  end_time   = "2026-09-02T04:00:00Z"
  comment    = "Monthly patching window"
}
