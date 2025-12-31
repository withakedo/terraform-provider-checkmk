# Host check interval - typed wrapper for extra_host_conf:check_interval
# This is a convenience resource that provides a typed interface

resource "checkmk_host_check_interval" "critical_hosts" {
  folder = "/"
  value  = 30 # seconds (integer)

  properties = {
    description = "30-second check interval for critical hosts"
  }

  conditions = {
    host_tags = [
      {
        key      = "criticality"
        operator = "is"
        value    = "critical"
      }
    ]
  }
}

# Example: Different interval for development hosts
resource "checkmk_host_check_interval" "dev_hosts" {
  folder = "/"
  value  = 300 # 5 minutes

  properties = {
    description = "5-minute check interval for development hosts"
    disabled    = false
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
