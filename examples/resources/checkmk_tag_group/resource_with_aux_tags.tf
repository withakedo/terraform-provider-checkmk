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

# Example with auxiliary tags
resource "checkmk_tag_group" "location" {
  id    = "location"
  title = "Data Center Location"
  topic = "Infrastructure"

  tags = [
    {
      id       = "dc1"
      title    = "Data Center 1 (US-East)"
      aux_tags = ["us", "east"]
    },
    {
      id       = "dc2"
      title    = "Data Center 2 (US-West)"
      aux_tags = ["us", "west"]
    },
    {
      id       = "dc3"
      title    = "Data Center 3 (EU-Central)"
      aux_tags = ["eu", "central"]
    },
  ]
}
