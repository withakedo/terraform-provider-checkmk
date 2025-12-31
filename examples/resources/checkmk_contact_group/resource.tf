# Basic contact group
resource "checkmk_contact_group" "admins" {
  name  = "admins"
  alias = "System Administrators"
}
