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

# Acknowledge a problem on a specific service, non-sticky and without notifying contacts
resource "checkmk_acknowledge" "cpu_load" {
  host_name           = checkmk_host.web.host_name
  service_description = "CPU load"
  comment             = "Known spike during nightly batch job"
  sticky              = false
  notify              = false
  expire_on           = "2026-09-02T06:00:00Z"
}
