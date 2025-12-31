# Basic user with password authentication
resource "checkmk_user" "operator" {
  username = "jsmith"
  fullname = "John Smith"
  email    = "jsmith@example.com"

  auth_type = "password"
  password  = "SecurePassword123!"
}
