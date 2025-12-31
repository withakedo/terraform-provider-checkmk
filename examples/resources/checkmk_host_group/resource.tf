# Basic host group
resource "checkmk_host_group" "webservers" {
  name  = "webservers"
  alias = "Web Servers"
}
