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

# List all active downtimes (host and service) on a host, e.g. to detect
# downtimes scheduled outside Terraform.
data "checkmk_downtimes" "web" {
  host_name = "web-01.example.com"
}

output "active_downtime_count" {
  value = length(data.checkmk_downtimes.web.downtimes)
}
