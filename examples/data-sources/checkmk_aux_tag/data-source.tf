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

# Look up an auxiliary tag by ID
data "checkmk_aux_tag" "example" {
  aux_tag_id = "production"
}

output "tag_title" {
  value = data.checkmk_aux_tag.example.title
}
