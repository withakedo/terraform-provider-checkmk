# Service retry interval - typed wrapper for extra_service_conf:retry_interval
# Controls how often to recheck a service in soft state

resource "checkmk_service_retry_interval" "critical_services" {
  folder = "/"
  value  = 15 # seconds

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

# Example: Standard retry interval for most services
resource "checkmk_service_retry_interval" "standard" {
  folder = "/"
  value  = 60 # 1 minute

  properties = {
    description = "Standard retry interval"
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
