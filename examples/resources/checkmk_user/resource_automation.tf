# Automation user for API access
resource "checkmk_user" "api_user" {
  username = "terraform_automation"
  fullname = "Terraform Automation User"

  auth_type         = "automation"
  automation_secret = "your-automation-secret-here"

  # Automation users typically have admin role
  roles = ["admin"]

  # Disable GUI login for automation users
  disable_login = true
}
