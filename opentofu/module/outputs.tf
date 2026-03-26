output "node_id" {
  value = wiredoor_node.this.id
}

output "node_token" {
  value     = wiredoor_node_pat.this.token
  sensitive = true
}

output "domain_id" {
  value = try(wiredoor_domain.this[0].id, null)
}

output "http_service_id" {
  value = wiredoor_http_service.this.id
}
