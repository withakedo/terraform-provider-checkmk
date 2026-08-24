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

# List all comments (host and service) on a host, e.g. to detect comments
# added outside Terraform.
data "checkmk_comments" "web" {
  host_name = "web-01.example.com"
}

output "comment_count" {
  value = length(data.checkmk_comments.web.comments)
}
