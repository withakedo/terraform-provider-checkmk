# Auxiliary tag used with tag groups
# Aux tags are automatically applied when a tag group choice is selected

resource "checkmk_aux_tag" "tcp_checks" {
  id    = "tcp_checks"
  title = "TCP Checks Enabled"
  topic = "Agent Type"
}

resource "checkmk_aux_tag" "snmp_checks" {
  id    = "snmp_checks"
  title = "SNMP Checks Enabled"
  topic = "Agent Type"
}

# Tag group that references aux tags
resource "checkmk_tag_group" "agent_type" {
  id    = "agent_type"
  title = "Agent Type"
  topic = "Monitoring"

  tags = [
    {
      id       = "cmk_agent"
      title    = "CheckMK Agent"
      aux_tags = [checkmk_aux_tag.tcp_checks.id]
    },
    {
      id       = "snmp_only"
      title    = "SNMP Only"
      aux_tags = [checkmk_aux_tag.snmp_checks.id]
    },
    {
      id    = "dual"
      title = "Agent + SNMP"
      aux_tags = [
        checkmk_aux_tag.tcp_checks.id,
        checkmk_aux_tag.snmp_checks.id
      ]
    },
  ]
}
