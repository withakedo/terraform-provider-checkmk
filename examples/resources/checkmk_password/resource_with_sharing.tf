# Password with ownership and sharing permissions
# Shows how to configure access control for stored credentials

resource "checkmk_contact_group" "network_ops" {
  name  = "network_ops"
  alias = "Network Operations Team"
}

resource "checkmk_contact_group" "cloud_team" {
  name  = "cloud_team"
  alias = "Cloud Infrastructure Team"
}

resource "checkmk_password" "network_credential" {
  password_id = "network_device_pwd"
  title       = "Network Device Password"
  password    = var.network_password

  # Owner of the password
  owner = "admin"

  # Contact groups that can edit the password
  editable_by = [checkmk_contact_group.network_ops.name]

  # Contact groups that can view/use the password
  share_with = [
    checkmk_contact_group.network_ops.name,
    checkmk_contact_group.cloud_team.name
  ]

  # Optional metadata
  documentation_url = "https://wiki.example.com/network-credentials"
  comment           = "Managed by Terraform - do not edit manually"

  depends_on = [
    checkmk_contact_group.network_ops,
    checkmk_contact_group.cloud_team
  ]
}

variable "network_password" {
  type      = string
  sensitive = true
}
