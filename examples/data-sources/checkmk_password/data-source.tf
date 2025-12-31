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

# Look up a password store entry by ID
# Note: The actual password value is never returned for security reasons
data "checkmk_password" "example" {
  id = "database_credentials"
}

output "password_title" {
  value = data.checkmk_password.example.title
}
