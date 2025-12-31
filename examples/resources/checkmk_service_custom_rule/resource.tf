# Service custom rule - typed wrapper for custom_checks ruleset
# Defines custom check commands for services

resource "checkmk_service_custom_rule" "app_health" {
  folder = "/"
  value  = "/usr/local/bin/check_app_health"

  properties = {
    description = "Custom application health check"
  }

  conditions = {
    host_tags = [
      {
        key      = "application"
        operator = "is"
        value    = "myapp"
      }
    ]
  }
}

# Example: Custom script with arguments
resource "checkmk_service_custom_rule" "database_check" {
  folder = "/"
  value  = "/usr/local/lib/nagios/plugins/check_mysql_health --mode connection-time"

  properties = {
    description = "MySQL health check"
    comment     = "Checks database connection time"
  }

  conditions = {
    host_tags = [
      {
        key      = "application"
        operator = "is"
        value    = "database"
      }
    ]
  }
}
