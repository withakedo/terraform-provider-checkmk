# Basic service group
resource "checkmk_service_group" "http_services" {
  name  = "http_services"
  alias = "HTTP/HTTPS Services"
}
