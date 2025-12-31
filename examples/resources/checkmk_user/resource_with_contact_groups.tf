# User with contact group memberships
resource "checkmk_contact_group" "linux_admins" {
  name  = "linux_admins"
  alias = "Linux Administrators"
}

resource "checkmk_contact_group" "database_admins" {
  name  = "database_admins"
  alias = "Database Administrators"
}

resource "checkmk_user" "dba" {
  username = "dba_alice"
  fullname = "Alice DBA"
  email    = "alice@example.com"

  contact_groups = [
    checkmk_contact_group.linux_admins.name,
    checkmk_contact_group.database_admins.name,
  ]

  roles = ["user"]

  auth_type = "password"
  password  = "DBAPassword123!"
}
