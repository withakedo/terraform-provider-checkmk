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

# Look up a service group by name
data "checkmk_service_group" "example" {
  name = "critical-services"
}

output "group_alias" {
  value = data.checkmk_service_group.example.alias
}
