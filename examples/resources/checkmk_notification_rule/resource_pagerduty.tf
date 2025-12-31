# PagerDuty escalation notification rule
resource "checkmk_notification_rule" "pagerduty_escalation" {
  description = "PagerDuty escalation for unacknowledged critical alerts"
  comment     = "Triggers after 15 minutes of unacknowledged critical state"

  contact_selection = {
    contact_groups = [checkmk_contact_group.tier2_oncall.name]
  }

  notification_method = {
    plugin_name = "pagerduty"
    plugin_params = jsonencode({
      routing_key = {
        password_store_id = checkmk_password.pagerduty_key.id
      }
    })
  }

  conditions = {
    match_host_event    = ["down"]
    match_service_event = ["crit"]

    match_host_labels = {
      criticality = "high"
    }

    # Only trigger after the first notification (escalation)
    restrict_notification_numbers = {
      beginning_from = 2
      up_to          = 999
    }
  }
}
