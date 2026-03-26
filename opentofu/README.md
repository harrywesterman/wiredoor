# Wiredoor OpenTofu Provider

This directory contains a native OpenTofu / Terraform provider for Wiredoor.

## What it manages

The provider currently supports:

- `wiredoor_node`
- `wiredoor_http_service`
- `wiredoor_tcp_service`
- `wiredoor_domain`
- `wiredoor_node_pat`
- `wiredoor_config` data source

## Build

```bash
cd opentofu/provider
go test ./...
go build
```

## Provider config

```hcl
provider "wiredoor" {
  endpoint = "https://wiredoor.example.com"
  username = var.username
  password = var.password
  insecure = false
}
```

You can also provide a pre-authenticated `token` instead of username/password.

## Import formats

- `wiredoor_node`: `<id>`
- `wiredoor_domain`: `<id>`
- `wiredoor_http_service`: `<node_id>/<service_id>`
- `wiredoor_tcp_service`: `<node_id>/<service_id>`
- `wiredoor_node_pat`: `<node_id>/<pat_id>`

## Module usage

The `opentofu/module` directory provides a small wrapper that creates:

- one Wiredoor node
- one node PAT
- one HTTP service
- an optional domain resource for SSL/certificate settings
- an optional TCP service

The basic example in `opentofu/examples/basic` shows how to wire the module into a root configuration.

The module defaults to `self-signed` domain SSL when `domain_name` is set. Set `domain_ssl = "certbot"` for public domains that should use Let's Encrypt.
Set `tcp_enabled = true` to create a TCP service alongside the HTTP service.

## Notes

- PAT tokens are treated as sensitive and are only persisted when the API returns them.
- This provider targets an existing Wiredoor deployment; it does not install Wiredoor itself.

See `opentofu/examples/basic` for a complete root module example.
