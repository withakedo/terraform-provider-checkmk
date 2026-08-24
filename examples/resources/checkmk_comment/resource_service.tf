terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"
}

resource "checkmk_host" "web" {
  host_name = "web-01.example.com"
  folder    = "/"

  attributes = {
    ipaddress = "10.1.2.3"
  }
}

# Comment on a specific service, persisted across CheckMK restarts
resource "checkmk_comment" "disk_note" {
  host_name           = checkmk_host.web.host_name
  service_description = "Disk /data"
  comment             = "Expanded to 2TB, thresholds still based on old size"
  persistent          = true
}
