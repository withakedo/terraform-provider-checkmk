terraform {
  required_providers {
    checkmk = {
      source = "withake-it/checkmk"
    }
  }
}

# Configure the CheckMK Provider
provider "checkmk" {
  # Configuration can be provided via environment variables:
  # export CHECKMK_URL="http://localhost:5000/test"
  # export CHECKMK_USERNAME="automation"
  # export CHECKMK_PASSWORD="your-secret"

  # Or explicitly in the provider block:
  url      = "http://localhost:5000/test"
  username = "automation"
  password = "your-secret-here"
}
