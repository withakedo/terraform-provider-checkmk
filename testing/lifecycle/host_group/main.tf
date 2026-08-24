terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

provider "checkmk" {}

resource "checkmk_host_group" "test" {
  name  = "lifecycle_test_hostgroup"
  alias = "Lifecycle Test Host Group"
}
