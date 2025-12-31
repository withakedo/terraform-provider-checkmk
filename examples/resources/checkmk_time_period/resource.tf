# Business hours time period
resource "checkmk_time_period" "business_hours" {
  name  = "business_hours"
  alias = "Business Hours (Mon-Fri 9-17)"

  active_time_ranges = [
    {
      day   = "monday"
      start = "09:00"
      end   = "17:00"
    },
    {
      day   = "tuesday"
      start = "09:00"
      end   = "17:00"
    },
    {
      day   = "wednesday"
      start = "09:00"
      end   = "17:00"
    },
    {
      day   = "thursday"
      start = "09:00"
      end   = "17:00"
    },
    {
      day   = "friday"
      start = "09:00"
      end   = "17:00"
    },
  ]
}
