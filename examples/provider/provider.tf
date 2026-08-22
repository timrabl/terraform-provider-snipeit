terraform {
  required_providers {
    snipeit = {
      source = "timrabl/snipeit"
    }
  }
}

provider "snipeit" {
  url   = "https://snipeit.example.com"
  token = var.snipeit_token # or via SNIPEIT_TOKEN environment variable

  # Only for dev instances with self-signed certificates:
  # insecure = true
}
