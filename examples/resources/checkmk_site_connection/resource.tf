terraform {
  required_providers {
    checkmk = {
      source = "withake-it/checkmk"
    }
  }
}

provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"
}

# Connects a remote monitoring site for distributed monitoring.
resource "checkmk_site_connection" "remote_dc2" {
  site_id = "remote_dc2"

  config_raw = jsonencode({
    basic_settings = {
      alias = "Remote datacenter 2"
    }
    configuration_connection = {
      enable_replication              = true
      url_of_remote_site              = "https://dc2.example.com/remote_dc2/check_mk/"
      disable_remote_configuration    = true
      ignore_tls_errors               = false
      direct_login_to_web_gui_allowed = true
      replicate_extensions            = true
      replicate_event_console         = true
    }
    status_connection = {
      connection = {
        socket_type = "tcp"
        host        = "dc2.example.com"
        port        = 6557
        encrypted   = true
        verify      = true
      }
      proxy = {
        use_livestatus_daemon = "direct"
      }
      connect_timeout       = 2
      persistent_connection = false
      disable_in_status_gui = false
    }
  })
}
