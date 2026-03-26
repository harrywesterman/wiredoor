terraform {
  required_providers {
    wiredoor = {
      source = "wiredoor/wiredoor"
    }
  }
}

resource "wiredoor_node" "this" {
  name           = var.node_name
  address        = var.node_addr
  allow_internet = var.allow_internet
  enabled        = var.node_enabled
}

resource "wiredoor_node_pat" "this" {
  node_id = wiredoor_node.this.id
  name    = var.pat_name
}

resource "wiredoor_http_service" "this" {
  node_id       = wiredoor_node.this.id
  name          = var.http_name
  path_location = var.http_path
  backend_port  = var.http_port
  backend_host  = var.http_backend_host
  backend_proto = var.http_backend_proto
  domain        = var.http_domain
  enabled       = var.http_enabled
  require_auth  = var.http_require_auth
}

