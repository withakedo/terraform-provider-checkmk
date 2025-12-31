# Basic folder example
resource "checkmk_folder" "webservers" {
  name  = "webservers"
  path  = "/"
  title = "Web Servers"
}

# Nested folder
resource "checkmk_folder" "webservers_prod" {
  name  = "production"
  path  = checkmk_folder.webservers.full_path
  title = "Production Web Servers"
}
