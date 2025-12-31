# Notification rule with comprehensive conditions
resource "checkmk_notification_rule" "business_hours_alerts" {
  description  = "Business hours alerts for web services"
  comment      = "Only notify during business hours for web-related services"
  allow_disable = true
  disabled      = false

  contact_selection = {
    specific_users = ["admin", "webops"]
    contact_groups = [
      checkmk_contact_group.web_team.name,
      checkmk_contact_group.devops.name
    ]
  }

  notification_method = {
    plugin_name   = "mail"
    plugin_params = jsonencode({
      from_address = "monitoring@example.com"
      subject      = "[ALERT] $HOSTNAME$ - $SERVICEDESC$"
    })
  }

  conditions = {
    # Event types to notify on
    match_host_event    = ["down", "up"]
    match_service_event = ["crit", "warn", "ok"]

    # Host filtering
    match_hostname = {
      match_type   = "regex"
      match_values = ["web-.*", "app-.*"]
    }

    match_host_tags = [
      {
        tag_group_id = "criticality"
        operator     = "is"
        tag_id       = "prod"
      }
    ]

    match_host_labels = {
      environment = "production"
      tier        = "frontend"
    }

    match_host_groups = [checkmk_host_group.webservers.name]

    # Service filtering
    match_service_description = {
      match_type   = "regex"
      match_values = ["HTTP.*", "HTTPS.*", "Certificate.*"]
    }

    match_service_groups = [checkmk_service_group.web_services.name]

    # Time-based filtering
    match_only_during_time_period = checkmk_time_period.business_hours.name

    # Folder filtering
    match_folder = checkmk_folder.production.path
  }
}
