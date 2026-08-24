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

# Look up a tag group by ID
data "checkmk_tag_group" "example" {
  tag_group_id = "environment"
}

output "group_title" {
  value = data.checkmk_tag_group.example.title
}
