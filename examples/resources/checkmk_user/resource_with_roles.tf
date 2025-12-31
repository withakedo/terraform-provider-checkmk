# User with specific roles
resource "checkmk_user" "admin_user" {
  username = "admin_jdoe"
  fullname = "Jane Doe (Admin)"
  email    = "jdoe@example.com"

  roles = ["admin"]

  auth_type = "password"
  password  = "AdminPassword123!"
}

resource "checkmk_user" "readonly_user" {
  username = "viewer_bob"
  fullname = "Bob View"
  email    = "bob@example.com"

  roles = ["guest"]

  auth_type = "password"
  password  = "ViewerPassword123!"
}
