# Service max check attempts - typed wrapper for extra_service_conf:max_check_attempts
# Controls how many times a service must fail before transitioning to hard state

resource "checkmk_service_max_check_attempts" "flaky_services" {
  folder = "/"
  value  = 5 # Number of attempts

  properties = {
    description = "5 retries before alerting on flaky network services"
  }

  conditions = {
    service_description = {
      match_on = ["Network*", "HTTP*"]
      operator = "one_of"
    }
  }
}

# Example: Fewer retries for critical services
resource "checkmk_service_max_check_attempts" "critical_services" {
  folder = "/"
  value  = 2 # Alert faster for critical services

  properties = {
    description = "Quick alerting for critical services"
  }

  conditions = {
    service_labels = [
      {
        key      = "priority"
        operator = "is"
        value    = "critical"
      }
    ]
  }
}
