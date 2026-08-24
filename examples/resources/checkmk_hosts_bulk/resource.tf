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

# Manages all three hosts with 1 API call per apply operation, instead of
# the 3 separate calls three checkmk_host resources with for_each would make.
resource "checkmk_hosts_bulk" "fleet" {
  host {
    host_name = "web-01.example.com"
    folder    = "/"
    attributes = {
      ipaddress = "10.1.2.1"
    }
  }

  host {
    host_name = "web-02.example.com"
    folder    = "/"
    attributes = {
      ipaddress = "10.1.2.2"
    }
  }

  host {
    host_name = "web-03.example.com"
    folder    = "/"
    attributes = {
      ipaddress = "10.1.2.3"
    }
  }

  activate = "auto"
}
