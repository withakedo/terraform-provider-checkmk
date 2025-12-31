# Multiple contact groups for different teams
resource "checkmk_contact_group" "network_team" {
  name  = "network_team"
  alias = "Network Operations Team"
}

resource "checkmk_contact_group" "database_team" {
  name  = "database_team"
  alias = "Database Administrators"
}

resource "checkmk_contact_group" "security_team" {
  name  = "security_team"
  alias = "Security Operations Center"
}

resource "checkmk_contact_group" "oncall" {
  name  = "oncall"
  alias = "On-Call Engineers"
}
