terraform {
  required_version = ">= 1.6.0"

  required_providers {
    wiredoor = {
      source  = "wiredoor/wiredoor"
      version = ">= 0.1.0"
    }
  }
}

provider "wiredoor" {
  endpoint = var.endpoint
  username = var.username
  password = var.password
  insecure = var.insecure
}

module "gateway" {
  source = "../../module"

  node_name  = var.node_name
  node_addr  = var.node_addr
  http_name  = var.http_name
  http_path  = var.http_path
  http_port  = var.http_port
}
