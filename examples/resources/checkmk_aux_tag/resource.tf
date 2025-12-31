# Basic auxiliary tag
resource "checkmk_aux_tag" "monitored_by_agent" {
  id    = "monitored_by_agent"
  title = "Monitored by Agent"
  topic = "Monitoring"
  help  = "Host is monitored via CheckMK agent"
}
