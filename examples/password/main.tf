terraform {
  required_providers {
    checkmk = {
      source = "blackmesaltd/checkmk"
    }
  }
}

provider "checkmk" {
  url      = "http://localhost:5030/test"
  username = "automation"
  password = var.checkmk_password
}

variable "checkmk_password" {
  type      = string
  sensitive = true
}

variable "azure_api_token" {
  type      = string
  sensitive = true
}

# Basic password example
resource "checkmk_password" "api_token" {
  password_id = "azure_api_token"
  title       = "Azure API Token"
  password    = var.azure_api_token
}

# Password with ownership and sharing
resource "checkmk_password" "shared_credential" {
  password_id      = "shared_db_password"
  title            = "Shared Database Password"
  password         = "supersecret123"
  owner            = "admin"
  documentation_url = "https://wiki.example.com/db-credentials"
  comment          = "Production database credentials"
}

# Password with contact group permissions
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
  password    = "networkpass456"
  owner       = "admin"
  editable_by = [checkmk_contact_group.network_ops.name]
  share_with  = [
    checkmk_contact_group.network_ops.name,
    checkmk_contact_group.cloud_team.name
  ]

  depends_on = [
    checkmk_contact_group.network_ops,
    checkmk_contact_group.cloud_team
  ]
}
