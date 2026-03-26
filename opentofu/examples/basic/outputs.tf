output "node_id" {
  value = module.gateway.node_id
}

output "node_token" {
  value     = module.gateway.node_token
  sensitive = true
}

output "http_service_id" {
  value = module.gateway.http_service_id
}

