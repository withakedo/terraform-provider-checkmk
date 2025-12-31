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

resource "checkmk_tag_group" "environment" {
  id    = "environment"
  title = "Environment"
  topic = "Custom Tags"
  help  = "Classify hosts by their environment"

  tags = [
    {
      id       = "prod"
      title    = "Production"
      aux_tags = []
    },
    {
      id       = "dev"
      title    = "Development"
      aux_tags = []
    },
    {
      id       = "test"
      title    = "Testing"
      aux_tags = []
    },
  ]
}
