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

# Look up a host group by name
data "checkmk_host_group" "example" {
  name = "linux-servers"
}

output "group_alias" {
  value = data.checkmk_host_group.example.alias
}
