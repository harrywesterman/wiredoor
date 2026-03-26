# Wiredoor OpenTofu Provider

This directory contains a native OpenTofu / Terraform provider for Wiredoor.

## Status

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

## Example usage

See `opentofu/examples/basic`.

