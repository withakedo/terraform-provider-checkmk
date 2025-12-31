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

# Example with resource-level strict ETag locking enabled
resource "checkmk_tag_group" "application" {
  id    = "application"
  title = "Application Type"

  # Enable strict resource locking for this specific resource
  # This will fetch and validate ETags to detect drift
  strict_resource_locking = true

  tags = [
    {
      id       = "web"
      title    = "Web Server"
      aux_tags = []
    },
    {
      id       = "db"
      title    = "Database Server"
      aux_tags = []
    },
    {
      id       = "app"
      title    = "Application Server"
      aux_tags = []
    },
  ]
}
