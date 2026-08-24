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

# Look up a contact group by name
data "checkmk_contact_group" "example" {
  name = "admins"
}

output "contact_group_alias" {
  value = data.checkmk_contact_group.example.alias
}
