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

# Look up a folder by path
data "checkmk_folder" "example" {
  path = "~production"
}

output "folder_title" {
  value = data.checkmk_folder.example.title
}
