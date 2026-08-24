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

# Aggregates the overall state of a host into a single BI health status.
resource "checkmk_bi_aggregation" "shop_frontend" {
  aggregation_id = "shop_frontend"
  pack_id        = "default"

  definition_raw = jsonencode({
    comment = "Overall health of the shop frontend host"
    groups = {
      names = ["Shop"]
    }
    node = {
      search = {
        type = "empty"
      }
      action = {
        type       = "state_of_host"
        host_regex = "shop-frontend.*"
      }
    }
    aggregation_visualization = {
      ignore_rule_styles = false
      layout_id          = "builtin_default"
      line_style         = "round"
    }
    computation_options = {
      disabled                               = false
      use_hard_states                        = false
      escalate_downtimes_as_warn             = false
      downtime_only_on_full_problem_coverage = false
      freeze_aggregations                    = false
    }
  })
}
