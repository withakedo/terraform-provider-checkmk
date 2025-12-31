# Basic rule example: Set host check interval
# This rule configures 60-second check intervals for production hosts

resource "checkmk_rule" "check_interval" {
  ruleset = "extra_host_conf:check_interval"
  folder  = "/"

  # Simple numeric values can be passed as strings
  # For the check_interval ruleset, the API expects a float
  value_raw = "60.0"

  properties = {
    description = "60-second check interval for production"
    comment     = "Managed by Terraform"
    disabled    = false
  }

  conditions = {
    host_tags = [
      {
        key      = "criticality"
        operator = "is"
        value    = "prod"
      }
    ]
  }
}
