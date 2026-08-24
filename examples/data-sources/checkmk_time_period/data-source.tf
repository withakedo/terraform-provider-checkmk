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

# Look up a time period by name
data "checkmk_time_period" "example" {
  name = "workhours"
}

output "time_period_alias" {
  value = data.checkmk_time_period.example.alias
}
