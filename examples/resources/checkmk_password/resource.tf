# Basic password example
# Stores a sensitive credential in the CheckMK password store

variable "azure_api_token" {
  type      = string
  sensitive = true
}

resource "checkmk_password" "api_token" {
  password_id = "azure_api_token"
  title       = "Azure API Token"
  password    = var.azure_api_token
}
