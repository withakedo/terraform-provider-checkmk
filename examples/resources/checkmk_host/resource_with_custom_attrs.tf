# Example: Using Custom Attributes
#
# Custom attributes are site-specific extensions that must be configured
# in CheckMK via CLI before they can be used in Terraform.
#
# See: docs/custom_attributes.md for setup instructions

provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "secret"

  activate = "auto"
}

# Example 1: Network device with custom attributes
resource "checkmk_host" "router" {
  host_name = "core-router-01"
  folder    = "/network/routers"

  attributes = {
    # Built-in attributes
    alias     = "Core Router 01"
    ipaddress = "10.0.1.1"
    site      = "datacenter-1"

    # Custom attributes (must be pre-configured in CheckMK)
    device_description = "Main datacenter core router"
    device_make        = "Cisco"
    proxy_port         = "8080"
    tcp_port           = "22"
  }
}

# Example 2: Server with ServiceNow integration
resource "checkmk_host" "app_server" {
  host_name = "app-prod-01"
  folder    = "/servers/production"

  attributes = {
    # Built-in attributes
    alias     = "Production Application Server"
    ipaddress = "10.0.2.50"

    # Custom attributes for ITSM integration
    device_description = "Production web application server"
    snowgroup          = "platform-team"   # ServiceNow assignment group
    snowservice        = "web-application" # ServiceNow service
  }
}

# Example 3: Using lifecycle.ignore_changes with custom attributes
resource "checkmk_host" "managed_device" {
  host_name = "network-switch-01"
  folder    = "/network/switches"

  attributes = {
    # Terraform manages these
    alias              = "Access Switch 01"
    ipaddress          = "10.0.3.100"
    device_make        = "Cisco"
    device_description = "Office access switch"

    # These can be modified via UI without causing drift
    notes      = "Initial configuration"
    proxy_port = "8080"
  }

  lifecycle {
    # Allow operators to update notes and proxy_port via UI
    ignore_changes = [
      attributes["notes"],
      attributes["proxy_port"],
    ]
  }
}

# Example 4: Error handling - undefined custom attribute
#
# ⚠️ This will FAIL if 'my_custom_field' is not configured in CheckMK
#
# resource "checkmk_host" "will_fail" {
#   host_name = "test-host"
#   folder    = "/"
#
#   attributes = {
#     alias            = "Test"
#     my_custom_field  = "value"  # ERROR: Unknown attribute
#   }
# }
#
# Error message:
#   Error: Client Error
#   Unable to create host, got error: API error (400):
#   Unknown attribute: 'my_custom_field'
#
# Solution: Configure the custom attribute in CheckMK first via CLI

# Example 5: Testing custom attributes
#
# Before deploying, verify custom attributes exist:
#
# curl -u "automation:password" \
#   "http://localhost:5000/test/check_mk/api/1.0/domain-types/host_config/collections/all" \
#   -X POST \
#   -H "Content-Type: application/json" \
#   -d '{
#     "host_name": "test-custom",
#     "folder": "/",
#     "attributes": {
#       "device_description": "Test"
#     }
#   }'

# Example 6: Dynamic custom attributes
#
# Custom attributes can be passed from variables or data sources
variable "device_metadata" {
  type = map(string)
  default = {
    device_make        = "Generic"
    device_description = "Standard configuration"
  }
}

resource "checkmk_host" "dynamic_attrs" {
  host_name = "dynamic-host"
  folder    = "/"

  attributes = merge(
    {
      alias     = "Dynamic Host"
      ipaddress = "10.0.4.100"
    },
    var.device_metadata # Merge in custom attributes from variable
  )
}
