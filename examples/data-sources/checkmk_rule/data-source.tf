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

# Look up a rule by its UUID
data "checkmk_rule" "example" {
  rule_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "rule_description" {
  value = data.checkmk_rule.example.description
}

output "rule_ruleset" {
  value = data.checkmk_rule.example.ruleset
}
