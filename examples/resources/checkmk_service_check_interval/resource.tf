# Service check interval - typed wrapper for extra_service_conf:check_interval
# This is a convenience resource that provides a typed interface

resource "checkmk_service_check_interval" "critical_services" {
  folder = "/"
  value  = 60 # seconds (integer)

  properties = {
    description = "60-second check interval for critical services"
  }

  conditions = {
    service_description = {
      match_on = ["CPU*", "Memory*"]
      operator = "one_of"
    }
  }
}

# Example: Longer interval for backup services
resource "checkmk_service_check_interval" "backup_services" {
  folder = "/"
  value  = 600 # 10 minutes

  properties = {
    description = "10-minute check interval for backup services"
  }

  conditions = {
    service_description = {
      match_on = ["Backup*"]
      operator = "one_of"
    }
    host_tags = [
      {
        key      = "criticality"
        operator = "is"
        value    = "prod"
      }
    ]
  }
}
