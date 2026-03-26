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

  node_name        = var.node_name
  node_addr        = var.node_addr
  domain_name      = var.domain_name
  domain_ssl       = var.domain_ssl
  http_name        = var.http_name
  http_path        = var.http_path
  http_port        = var.http_port
  tcp_enabled      = var.tcp_enabled
  tcp_name         = var.tcp_name
  tcp_port         = var.tcp_port
  tcp_backend_port = var.tcp_backend_port
}
