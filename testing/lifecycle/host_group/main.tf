terraform {
  required_providers {
    checkmk = {
      source = "blackmesaltd/checkmk"
    }
  }
}

provider "checkmk" {}

resource "checkmk_host_group" "test" {
  name  = "lifecycle_test_hostgroup"
  alias = "Lifecycle Test Host Group"
}
