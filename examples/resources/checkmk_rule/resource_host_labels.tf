# Host label assignment rule
# Assigns labels to hosts based on host name patterns
#
# IMPORTANT: host_label_rules CANNOT use host_labels conditions
# (would create circular dependency)
#
# NOTE: Requires Terraform 1.8+ for provider::checkmk::topython() function

resource "checkmk_rule" "production_labels" {
  ruleset = "host_label_rules"
  folder  = "/"

  # Use topython() to convert HCL maps to Python literal format
  # CheckMK expects: {'environment': 'production', ...}
  # NOT JSON: {"environment": "production", ...}
  value_raw = provider::checkmk::topython({
    environment = "production"
    tier        = "frontend"
    managed_by  = "terraform"
  })

  properties = {
    description = "Label production web servers"
    comment     = "Applied to web-* hosts"
    disabled    = false
  }

  conditions = {
    # Match hosts by name pattern
    host_name = {
      match_on = ["web-*", "app-*"]
      operator = "one_of"
    }

    # Match by host tags (allowed on label rulesets)
    host_tags = [
      {
        key      = "site"
        operator = "is"
        value    = "datacenter_a"
      }
    ]

    # NOTE: Cannot use host_labels here - would cause circular dependency
  }
}

# Another example: Label database servers
resource "checkmk_rule" "database_labels" {
  ruleset = "host_label_rules"
  folder  = "/"

  value_raw = provider::checkmk::topython({
    role       = "database"
    backup     = "required"
    compliance = "pci-dss"
  })

  properties = {
    description = "Label database servers"
  }

  conditions = {
    host_name = {
      match_on = ["db-*", "mysql-*", "postgres-*"]
      operator = "one_of"
    }
  }
}

# Example using variables with topython()
variable "default_labels" {
  type = map(string)
  default = {
    source = "terraform"
  }
}

resource "checkmk_rule" "default_labels" {
  ruleset = "host_label_rules"
  folder  = "/"

  # Variables work seamlessly with topython()
  value_raw = provider::checkmk::topython(var.default_labels)

  properties = {
    description = "Default labels for all hosts"
  }
}
