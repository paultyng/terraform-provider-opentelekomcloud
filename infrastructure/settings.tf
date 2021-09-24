terraform {
  required_version = ">= 0.13"

  required_providers {
    opentelekomcloud = {
      source  = "opentelekomcloud/opentelekomcloud"
      version = ">= 1.23.9"
    }
  }
}

provider "opentelekomcloud" {}
