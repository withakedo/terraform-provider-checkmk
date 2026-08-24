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

# Look up a host by name
data "checkmk_host" "example" {
  host_name = "my-server"
}

output "host_folder" {
  value = data.checkmk_host.example.folder
}
