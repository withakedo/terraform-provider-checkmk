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

# Look up a user by username
data "checkmk_user" "example" {
  username = "admin"
}

output "user_fullname" {
  value = data.checkmk_user.example.fullname
}

output "user_roles" {
  value = data.checkmk_user.example.roles
}
