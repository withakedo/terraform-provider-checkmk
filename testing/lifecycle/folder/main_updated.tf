terraform {
  required_providers {
    checkmk = {
      source = "withakedo/checkmk"
    }
  }
}

provider "checkmk" {}

resource "checkmk_folder" "test" {
  name  = "lifecycle_test_folder"
  path  = "/"
  title = "Lifecycle Test Folder - Updated"
}
