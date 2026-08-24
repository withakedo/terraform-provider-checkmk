terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

provider "checkmk" {}

resource "checkmk_user" "test" {
  username = "lifecycle_test_user"
  fullname = "Lifecycle Test User - Updated"

  auth_type = "password"
  password  = "TestPassword123!"
}
