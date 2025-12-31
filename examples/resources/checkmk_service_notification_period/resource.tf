# Service notification period - typed wrapper for extra_service_conf:notification_period
# This is a convenience resource that provides a typed interface

resource "checkmk_service_notification_period" "critical_services" {
  folder = "/"
  value  = "24X7" # Reference to an existing time period

  properties = {
    description = "24x7 notifications for critical services"
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

# Example: Quiet hours for non-critical services
resource "checkmk_service_notification_period" "non_critical" {
  folder = "/"
  value  = "workhours"

  properties = {
    description = "Business hours notifications for non-critical services"
  }

  conditions = {
    service_description = {
      match_on = ["Backup*", "Log*"]
      operator = "one_of"
    }
  }
}
