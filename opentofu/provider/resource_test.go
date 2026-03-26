package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testResourceData(t *testing.T, r *schema.Resource, values map[string]interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, r.Schema, values)
}

func TestNodePayloadMapping(t *testing.T) {
	t.Parallel()

	d := testResourceData(t, resourceNode(), map[string]interface{}{
		"name":             "node-a",
		"dns":              "1.1.1.1",
		"keepalive":        25,
		"address":          "10.0.0.2",
		"allow_internet":   true,
		"advanced":         true,
		"enabled":          false,
		"is_gateway":       true,
		"gateway_networks": []interface{}{map[string]interface{}{"interface": "eth1", "subnet": "10.1.0.0/24"}},
	})

	payload := nodePayload(d)

	if payload["name"] != "node-a" || payload["dns"] != "1.1.1.1" {
		t.Fatalf("unexpected node payload: %#v", payload)
	}
	if payload["allowInternet"] != true || payload["isGateway"] != true {
		t.Fatalf("unexpected node payload flags: %#v", payload)
	}
}

func TestConfigDataSourceRead(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/config" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"VPN_HOST":                "vpn.example.com",
			"TCP_SERVICES_PORT_RANGE": "32760-32767",
		})
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Token: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	d := testResourceData(t, dataSourceConfig(), map[string]interface{}{})
	diags := dataSourceConfigRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("read config: %v", diags)
	}

	if got := d.Get("vpn_host").(string); got != "vpn.example.com" {
		t.Fatalf("unexpected vpn host: %q", got)
	}
	if got := d.Get("tcp_services_port_range").(string); got != "32760-32767" {
		t.Fatalf("unexpected port range: %q", got)
	}
}

func TestCompositeImportRejectsInvalidIDs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "http",
			fn: func() error {
				_, err := resourceHTTPServiceImport(context.Background(), testResourceData(t, resourceHTTPService(), map[string]interface{}{}), nil)
				return err
			},
		},
		{
			name: "tcp",
			fn: func() error {
				_, err := resourceTCPServiceImport(context.Background(), testResourceData(t, resourceTCPService(), map[string]interface{}{}), nil)
				return err
			},
		},
		{
			name: "pat",
			fn: func() error {
				_, err := resourcePATImport(context.Background(), testResourceData(t, resourcePAT(), map[string]interface{}{}), nil)
				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.fn()
			if err == nil {
				t.Fatal("expected import error")
			}
		})
	}
}

func TestNodeReadRefreshClearsDerivedValues(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/nodes/7" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              7,
			"name":            "node-a",
			"address":         "10.0.0.2",
			"enabled":         true,
			"advanced":        false,
			"isGateway":       true,
			"gatewayNetworks": []map[string]interface{}{},
		})
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Token: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	d := testResourceData(t, resourceNode(), map[string]interface{}{
		"name":             "node-a",
		"address":          "10.0.0.2",
		"enabled":          true,
		"is_gateway":       true,
		"gateway_networks": []interface{}{map[string]interface{}{"interface": "eth1", "subnet": "10.1.0.0/24"}},
		"token":            "stale-token",
	})
	d.SetId("7")

	diags := resourceNodeRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("read node: %v", diags)
	}

	if got := d.Get("token").(string); got != "" {
		t.Fatalf("expected token to be cleared, got %q", got)
	}
	if got := d.Get("gateway_networks").([]interface{}); len(got) != 0 {
		t.Fatalf("expected gateway networks to be cleared, got %#v", got)
	}
}

func TestHTTPServicePayloadAndImport(t *testing.T) {
	t.Parallel()

	d := testResourceData(t, resourceHTTPService(), map[string]interface{}{
		"name":             "app",
		"domain":           "app.example.com",
		"path_location":    "/",
		"backend_host":     "10.0.0.10",
		"backend_port":     3000,
		"backend_proto":    "https",
		"allowed_ips":      []interface{}{"10.0.0.0/24"},
		"blocked_ips":      []interface{}{"10.2.0.0/24"},
		"require_auth":     true,
		"skip_auth_routes": "/health",
		"enabled":          false,
		"ttl":              "2h",
	})

	payload := httpServicePayload(d)
	if payload["backendHost"] != "10.0.0.10" || payload["backendProto"] != "https" {
		t.Fatalf("unexpected http payload: %#v", payload)
	}
	if _, ok := payload["allowedIps"]; !ok {
		t.Fatalf("missing allowedIps in payload: %#v", payload)
	}

	d.SetId("9/7")
	imported, err := resourceHTTPServiceImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("import http service: %v", err)
	}
	if len(imported) != 1 || imported[0].Id() != "9/7" {
		t.Fatalf("unexpected import result: %#v", imported)
	}
}

func TestTCPServicePayloadAndImport(t *testing.T) {
	t.Parallel()

	d := testResourceData(t, resourceTCPService(), map[string]interface{}{
		"name":         "stream",
		"domain":       "tcp.example.com",
		"proto":        "udp",
		"backend_host": "10.0.0.11",
		"backend_port": 9000,
		"port":         40001,
		"ssl":          true,
		"allowed_ips":  []interface{}{"10.0.0.0/24"},
		"blocked_ips":  []interface{}{"10.2.0.0/24"},
		"enabled":      true,
		"ttl":          "1h",
	})

	payload := tcpServicePayload(d)
	if payload["proto"] != "udp" || payload["backendPort"] != 9000 {
		t.Fatalf("unexpected tcp payload: %#v", payload)
	}

	d.SetId("5/11")
	imported, err := resourceTCPServiceImport(context.Background(), d, nil)
	if err != nil {
		t.Fatalf("import tcp service: %v", err)
	}
	if len(imported) != 1 || imported[0].Id() != "5/11" {
		t.Fatalf("unexpected import result: %#v", imported)
	}
}

func TestDomainPayloadAndImport(t *testing.T) {
	t.Parallel()

	d := testResourceData(t, resourceDomain(), map[string]interface{}{
		"domain":          "app.example.com",
		"ssl":             "certbot",
		"authentication":  true,
		"allowed_emails":  []interface{}{"ops@example.com"},
		"skip_validation": false,
	})

	payload := domainPayload(d)
	if payload["ssl"] != "certbot" || payload["authentication"] != true {
		t.Fatalf("unexpected domain payload: %#v", payload)
	}
}

func TestPATImportAndReadRefresh(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/nodes/9/pats" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "nodeId": 9, "name": "default", "revoked": false},
			{"id": 7, "nodeId": 9, "name": "deploy", "revoked": false},
		})
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Token: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	d := testResourceData(t, resourcePAT(), map[string]interface{}{
		"node_id": "9",
		"name":    "deploy",
		"token":   "stale-token",
	})
	d.SetId("9/7")

	if _, err := resourcePATImport(context.Background(), d, nil); err != nil {
		t.Fatalf("import pat: %v", err)
	}

	diags := resourcePATRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("read pat: %v", diags)
	}

	if d.Id() != "9/7" {
		t.Fatalf("unexpected id: %s", d.Id())
	}
	if got := d.Get("token").(string); got != "" {
		t.Fatalf("expected token to be cleared, got %q", got)
	}
	if got := d.Get("name").(string); got != "deploy" {
		t.Fatalf("unexpected name: %q", got)
	}
}

func TestHTTPServiceReadRefresh(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/services/9/http/7" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             7,
			"nodeId":         9,
			"name":           "app",
			"domain":         "app.example.com",
			"pathLocation":   "/",
			"backendHost":    "10.0.0.10",
			"backendPort":    3000,
			"backendProto":   "http",
			"requireAuth":    true,
			"skipAuthRoutes": "/health",
			"enabled":        true,
			"publicAccess":   "https://app.example.com/",
		})
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Token: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	d := testResourceData(t, resourceHTTPService(), map[string]interface{}{
		"node_id":          "9",
		"name":             "app",
		"domain":           "app.example.com",
		"path_location":    "/",
		"backend_host":     "10.0.0.10",
		"backend_port":     3000,
		"backend_proto":    "http",
		"require_auth":     true,
		"skip_auth_routes": "/health",
		"enabled":          true,
	})
	d.SetId("9/7")

	diags := resourceHTTPServiceRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("read http service: %v", diags)
	}

	if d.Id() != "9/7" {
		t.Fatalf("unexpected id: %s", d.Id())
	}
	if got := d.Get("public_access").(string); got != "https://app.example.com/" {
		t.Fatalf("unexpected public access: %q", got)
	}
}

func TestTCPServiceReadRefresh(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/services/5/tcp/11" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           11,
			"nodeId":       5,
			"name":         "stream",
			"domain":       "tcp.example.com",
			"proto":        "udp",
			"backendHost":  "10.0.0.11",
			"backendPort":  9000,
			"port":         40001,
			"ssl":          true,
			"enabled":      true,
			"publicAccess": "udp://tcp.example.com:40001",
		})
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Token: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	d := testResourceData(t, resourceTCPService(), map[string]interface{}{
		"node_id":      "5",
		"name":         "stream",
		"domain":       "tcp.example.com",
		"proto":        "udp",
		"backend_host": "10.0.0.11",
		"backend_port": 9000,
		"port":         40001,
		"ssl":          true,
		"enabled":      true,
	})
	d.SetId("5/11")

	diags := resourceTCPServiceRead(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("read tcp service: %v", diags)
	}

	if d.Id() != "5/11" {
		t.Fatalf("unexpected id: %s", d.Id())
	}
	if got := d.Get("public_access").(string); got != "udp://tcp.example.com:40001" {
		t.Fatalf("unexpected public access: %q", got)
	}
}
