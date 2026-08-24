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

# Look up a notification rule by its UUID
data "checkmk_notification_rule" "example" {
  rule_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "notification_rule_title" {
  value = data.checkmk_notification_rule.example.title
}
