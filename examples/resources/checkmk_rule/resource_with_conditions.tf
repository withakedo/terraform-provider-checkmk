# Rule with multiple condition types
# Demonstrates combining different condition types (all conditions are AND'ed)

resource "checkmk_rule" "critical_disk_checks" {
  ruleset = "extra_service_conf:check_interval"
  folder  = "/"

  # Check every 30 seconds
  value_raw = "30.0"

  properties = {
    description = "Fast disk checks for critical production systems"
    comment     = "Ensures disk issues are detected quickly"
    disabled    = false
  }

  conditions = {
    # Match specific hosts by name pattern
    host_name = {
      match_on = ["prod-*", "critical-*"]
      operator = "one_of"
    }

    # Must have production tag
    host_tags = [
      {
        key      = "criticality"
        operator = "is"
        value    = "prod"
      },
      {
        key      = "site"
        operator = "is_not"
        value    = "dr"
      }
    ]

    # Must have specific labels
    host_labels = [
      {
        key      = "environment"
        operator = "is"
        value    = "production"
      }
    ]

    # Only for disk/filesystem services
    service_description = {
      match_on = ["Disk*", "Filesystem*", "Mount*"]
      operator = "one_of"
    }
  }
}

# Example: Exclude certain hosts from monitoring
resource "checkmk_rule" "exclude_test_hosts" {
  ruleset = "extra_host_conf:check_interval"
  folder  = "/"

  # Slow check interval for test hosts (5 minutes)
  value_raw = "300.0"

  properties = {
    description = "Slow checks for test environment"
  }

  conditions = {
    host_name = {
      match_on = ["test-*", "dev-*", "sandbox-*"]
      operator = "one_of"
    }

    host_tags = [
      {
        key      = "criticality"
        operator = "is_not"
        value    = "prod"
      }
    ]
  }
}
