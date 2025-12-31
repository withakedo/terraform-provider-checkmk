# Multiple service groups for service categorization
resource "checkmk_service_group" "database_services" {
  name  = "database_services"
  alias = "Database Services (MySQL, PostgreSQL, etc.)"
}

resource "checkmk_service_group" "filesystem" {
  name  = "filesystem"
  alias = "Filesystem and Disk Services"
}

resource "checkmk_service_group" "memory_cpu" {
  name  = "memory_cpu"
  alias = "Memory and CPU Services"
}

resource "checkmk_service_group" "network_services" {
  name  = "network_services"
  alias = "Network Interface Services"
}

resource "checkmk_service_group" "backup_services" {
  name  = "backup_services"
  alias = "Backup and Recovery Services"
}
