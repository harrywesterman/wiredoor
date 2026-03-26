terraform {
  required_providers {
    wiredoor = {
      source = "wiredoor/wiredoor"
    }
  }
}

locals {
  http_domain = var.http_domain != "" ? var.http_domain : var.domain_name
  tcp_domain  = var.tcp_domain != "" ? var.tcp_domain : var.domain_name
  tcp_port    = var.tcp_port > 0 ? var.tcp_port : null
}

resource "wiredoor_node" "this" {
  name           = var.node_name
  address        = var.node_addr
  allow_internet = var.allow_internet
  enabled        = var.node_enabled
}

resource "wiredoor_domain" "this" {
  count = var.domain_name != "" ? 1 : 0

  domain          = var.domain_name
  ssl             = var.domain_ssl
  authentication   = var.domain_authentication
  allowed_emails   = var.domain_allowed_emails
  skip_validation  = var.domain_skip_validation
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
  domain        = local.http_domain
  enabled       = var.http_enabled
  require_auth  = var.http_require_auth
}

resource "wiredoor_tcp_service" "this" {
  count = var.tcp_enabled ? 1 : 0

  node_id      = wiredoor_node.this.id
  name         = var.tcp_name
  domain       = local.tcp_domain
  proto        = var.tcp_proto
  backend_host = var.tcp_backend_host
  backend_port = var.tcp_backend_port
  port         = local.tcp_port
  ssl          = var.tcp_ssl
  allowed_ips  = var.tcp_allowed_ips
  blocked_ips  = var.tcp_blocked_ips
  enabled      = var.tcp_enabled
  ttl          = var.tcp_ttl
}
