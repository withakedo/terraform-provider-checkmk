# Host notification period - typed wrapper for extra_host_conf:notification_period
# This is a convenience resource that provides a typed interface

resource "checkmk_host_notification_period" "production" {
  folder = "/"
  value  = "24X7" # Reference to an existing time period

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

# Example: Business hours only for development
resource "checkmk_host_notification_period" "development" {
  folder = "/"
  value  = "workhours"

  properties = {
    description = "Business hours notifications for dev hosts"
  }

  conditions = {
    host_tags = [
      {
        key      = "criticality"
        operator = "is"
        value    = "test"
      }
    ]
  }
}
