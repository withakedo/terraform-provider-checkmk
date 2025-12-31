# Examples of common ruleset types
# Shows how to use different rulesets for various monitoring configurations
#
# NOTE: Simple scalar values (numbers, strings) can be passed directly.
# Use provider::checkmk::topython() for complex values (maps, lists).
# Requires Terraform 1.8+ for provider functions.

# Service check interval - control how often services are checked
resource "checkmk_rule" "service_check_interval" {
  ruleset   = "extra_service_conf:check_interval"
  folder    = "/"
  value_raw = "120.0" # 2 minutes (float in seconds)

  properties = {
    description = "2-minute service check interval"
  }

  conditions = {
    service_description = {
      match_on = ["Memory", "CPU*"]
      operator = "one_of"
    }
  }
}

# Host notification period - when to send host alerts
# Value is a reference to a time period name
resource "checkmk_rule" "host_notification_period" {
  ruleset   = "extra_host_conf:notification_period"
  folder    = "/"
  value_raw = "'24X7'" # String value needs Python quotes

  properties = {
    description = "24x7 notifications for production hosts"
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

# Service notification period
resource "checkmk_rule" "service_notification_period" {
  ruleset   = "extra_service_conf:notification_period"
  folder    = "/"
  value_raw = "'workhours'" # Reference to a time period

  properties = {
    description = "Business hours notifications for non-critical services"
  }

  conditions = {
    host_tags = [
      {
        key      = "criticality"
        operator = "is_not"
        value    = "prod"
      }
    ]
  }
}

# Max check attempts - how many failures before alerting
resource "checkmk_rule" "max_check_attempts" {
  ruleset   = "extra_service_conf:max_check_attempts"
  folder    = "/"
  value_raw = "3"

  properties = {
    description = "3 retries before alerting on flaky services"
  }

  conditions = {
    service_description = {
      match_on = ["Network*", "HTTP*"]
      operator = "one_of"
    }
  }
}

# Retry interval - time between retries in soft state
resource "checkmk_rule" "retry_interval" {
  ruleset   = "extra_service_conf:retry_interval"
  folder    = "/"
  value_raw = "30.0" # 30 seconds between retries

  properties = {
    description = "Fast retries for critical services"
  }

  conditions = {
    service_labels = [
      {
        key      = "priority"
        operator = "is"
        value    = "high"
      }
    ]
  }
}

# Example with complex value using topython()
# Custom check parameters that require a dict/map value
resource "checkmk_rule" "filesystem_thresholds" {
  ruleset = "checkgroup_parameters:filesystem"
  folder  = "/"

  # Complex rule values require topython()
  value_raw = provider::checkmk::topython({
    levels          = [80.0, 90.0]
    magic_normsize  = 20
    levels_low      = [50.0, 60.0]
    trend_range     = 24
    trend_perfdata  = true
    show_levels     = "onmagic"
    inodes_levels   = [10.0, 5.0]
    show_inodes     = "onlow"
    show_reserved   = false
  })

  properties = {
    description = "Custom filesystem thresholds"
  }

  conditions = {
    service_description = {
      match_on = ["Filesystem*", "Mount*"]
      operator = "one_of"
    }
  }
}
