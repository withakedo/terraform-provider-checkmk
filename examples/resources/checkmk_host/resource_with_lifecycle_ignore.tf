# Example: Using lifecycle.ignore_changes for Click-Ops Attributes
#
# Some attributes in CheckMK are commonly modified through the UI ("click-ops")
# rather than through infrastructure-as-code. Use Terraform's built-in
# lifecycle.ignore_changes to prevent drift detection for these attributes.
#
# Common click-ops attributes:
# - site: CheckMK site assignment (often changed during ops)
# - snmp_community: SNMP community strings (security-sensitive)
# - tag_snmp_ds: SNMP data source tags
# - tag_agent: Agent type tags
#
# This allows Terraform to manage core host configuration while allowing
# operators to modify operational attributes through the UI.

provider "checkmk" {
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "secret"

  activate = "auto"
}

resource "checkmk_host" "managed_host" {
  host_name = "production-server"
  folder    = "/production"

  # Terraform manages these core attributes
  attributes = {
    alias     = "Production Application Server"
    ipaddress = "10.0.1.50"

    # Initial values for click-ops attributes
    # These can be changed in the UI without causing drift
    site           = "site1"
    snmp_community = "public"
    tag_snmp_ds    = "snmp-v2"
  }

  # Ignore changes to operational attributes that may be modified via UI
  lifecycle {
    ignore_changes = [
      attributes["site"],
      attributes["snmp_community"],
      attributes["tag_snmp_ds"],
    ]
  }
}

# Example: Partial ignore - manage some attributes, ignore others
resource "checkmk_host" "mixed_management" {
  host_name = "database-server"
  folder    = "/databases"

  attributes = {
    alias           = "Production Database"
    ipaddress       = "10.0.2.100"
    site            = "site1"
    snmp_community  = "public"
    tag_criticality = "prod"
  }

  lifecycle {
    # Only ignore operational attributes
    # Terraform will still detect drift for alias, ipaddress, tag_criticality
    ignore_changes = [
      attributes["site"],
      attributes["snmp_community"],
    ]
  }
}

# Example: Ignore all attributes (Terraform only manages host existence)
resource "checkmk_host" "minimal_management" {
  host_name = "testing-server"
  folder    = "/testing"

  attributes = {
    alias     = "Testing Server"
    ipaddress = "10.0.3.50"
  }

  lifecycle {
    # Ignore all attribute changes - only manage host existence and folder
    ignore_changes = [attributes]
  }
}

# Example: No lifecycle block - strict management
resource "checkmk_host" "strict_management" {
  host_name = "critical-server"
  folder    = "/critical"

  attributes = {
    alias     = "Critical Infrastructure"
    ipaddress = "10.0.4.100"
  }

  # No lifecycle block - any UI changes will be detected as drift
  # and reconciled on next terraform apply
  # Recommended for compliance scenarios with strict_resource_locking
}
