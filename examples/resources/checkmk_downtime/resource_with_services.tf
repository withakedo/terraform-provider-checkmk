terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

resource "checkmk_host" "db" {
  host_name = "db-01.example.com"
  folder    = "/"

  attributes = {
    ipaddress = "10.1.2.4"
  }
}

# Service-specific downtime
resource "checkmk_downtime" "db_backup" {
  host_name            = checkmk_host.db.host_name
  service_descriptions = ["Backup Job", "Disk IO"]
  start_time           = "2026-09-01T01:00:00Z"
  end_time             = "2026-09-01T03:00:00Z"
  comment              = "Nightly backup, expected high disk IO"
}
