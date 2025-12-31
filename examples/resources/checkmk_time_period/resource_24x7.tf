# 24x7 time period (always active)
resource "checkmk_time_period" "always" {
  name  = "always_on"
  alias = "24x7 - Always Active"

  active_time_ranges = [
    { day = "monday", start = "00:00", end = "24:00" },
    { day = "tuesday", start = "00:00", end = "24:00" },
    { day = "wednesday", start = "00:00", end = "24:00" },
    { day = "thursday", start = "00:00", end = "24:00" },
    { day = "friday", start = "00:00", end = "24:00" },
    { day = "saturday", start = "00:00", end = "24:00" },
    { day = "sunday", start = "00:00", end = "24:00" },
  ]
}
