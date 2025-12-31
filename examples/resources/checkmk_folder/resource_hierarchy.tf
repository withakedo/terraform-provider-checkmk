# Complete folder hierarchy example
resource "checkmk_folder" "datacenter" {
  name  = "datacenter"
  path  = "/"
  title = "Main Datacenter"
}

resource "checkmk_folder" "dc_network" {
  name  = "network"
  path  = checkmk_folder.datacenter.full_path
  title = "Network Infrastructure"
}

resource "checkmk_folder" "dc_servers" {
  name  = "servers"
  path  = checkmk_folder.datacenter.full_path
  title = "Server Infrastructure"
}

resource "checkmk_folder" "dc_servers_linux" {
  name  = "linux"
  path  = checkmk_folder.dc_servers.full_path
  title = "Linux Servers"
}

resource "checkmk_folder" "dc_servers_windows" {
  name  = "windows"
  path  = checkmk_folder.dc_servers.full_path
  title = "Windows Servers"
}
