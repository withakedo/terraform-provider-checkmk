terraform {
  required_providers {
    checkmk = {
      source = "withake-it/checkmk"
    }
  }
}

provider "checkmk" {}

resource "checkmk_user" "test" {
  username = "lifecycle_test_user"
  fullname = "Lifecycle Test User"

  auth_type = "password"
  password  = "TestPassword123!"
}
