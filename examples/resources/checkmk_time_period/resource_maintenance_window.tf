# Maintenance window - Sunday 2-6 AM
resource "checkmk_time_period" "maintenance_window" {
  name  = "maintenance_window"
  alias = "Weekly Maintenance Window"

  active_time_ranges = [
    {
      day   = "sunday"
      start = "02:00"
      end   = "06:00"
    },
  ]
}

# Extended maintenance window - weekends
resource "checkmk_time_period" "weekend_maintenance" {
  name  = "weekend_maintenance"
  alias = "Weekend Maintenance"

  active_time_ranges = [
    { day = "saturday", start = "00:00", end = "24:00" },
    { day = "sunday", start = "00:00", end = "24:00" },
  ]
}
