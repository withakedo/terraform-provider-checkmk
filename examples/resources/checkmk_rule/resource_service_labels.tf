# Service label assignment rule
# Assigns labels to services based on service description patterns
#
# IMPORTANT: service_label_rules CANNOT use service_labels conditions
# (would create circular dependency)
#
# NOTE: Requires Terraform 1.8+ for provider::checkmk::topython() function

resource "checkmk_rule" "database_service_labels" {
  ruleset = "service_label_rules"
  folder  = "/"

  # Use topython() to convert HCL maps to Python literal format
  value_raw = provider::checkmk::topython({
    service_type = "database"
    monitoring   = "critical"
  })

  properties = {
    description = "Label database services"
  }

  conditions = {
    # Match services by description pattern
    service_description = {
      match_on = ["MySQL*", "PostgreSQL*", "Oracle*", "MSSQL*"]
      operator = "one_of"
    }

    # Can still use host_labels (not circular for service labels)
    host_labels = [
      {
        key      = "role"
        operator = "is"
        value    = "database"
      }
    ]

    # NOTE: Cannot use service_labels here - would cause circular dependency
  }
}
