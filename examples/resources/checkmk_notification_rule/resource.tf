# Basic email notification rule
resource "checkmk_notification_rule" "email_alerts" {
  description = "Email alerts for critical events"

  contact_selection = {
    all_contacts_of_object = true
  }

  notification_method = {
    plugin_name = "mail"
  }

  conditions = {
    match_service_event = ["crit"]
  }
}
