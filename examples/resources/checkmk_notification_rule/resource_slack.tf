# Slack notification rule for critical alerts
resource "checkmk_notification_rule" "slack_critical" {
  description = "Slack notification for critical alerts"
  comment     = "Managed by Terraform"

  contact_selection = {
    contact_groups = [checkmk_contact_group.ops.name]
  }

  notification_method = {
    plugin_name   = "slack"
    plugin_params = jsonencode({
      webhook_url = {
        password_store_id = checkmk_password.slack_webhook.id
      }
      channel = "#alerts-critical"
    })
  }

  conditions = {
    match_host_event    = ["down", "unreachable"]
    match_service_event = ["crit"]
    match_host_labels = {
      environment = "production"
    }
  }
}
