# Folder with custom attributes
resource "checkmk_folder" "databases" {
  name  = "databases"
  path  = "/"
  title = "Database Servers"

  attributes = {
    tag_criticality = "prod"
    tag_networking  = "lan"
  }
}
