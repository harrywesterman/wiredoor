package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientAuthenticateAndUseToken(t *testing.T) {
	t.Parallel()

	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/login":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"abc123","expiresIn":"1h"}`))
		case "/nodes":
			authHeader = r.Header.Get("Authorization")
			body, _ := io.ReadAll(r.Body)
			var got map[string]interface{}
			_ = json.Unmarshal(body, &got)

			if got["name"] != "node-a" {
				t.Fatalf("unexpected node payload: %#v", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":7,"name":"node-a","address":"10.0.0.2","enabled":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Username: "admin", Password: "secret"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if err := client.Authenticate("admin", "secret"); err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if _, err := client.CreateNode(map[string]interface{}{"name": "node-a"}); err != nil {
		t.Fatalf("create node: %v", err)
	}

	if authHeader != "Bearer abc123" {
		t.Fatalf("expected bearer auth header, got %q", authHeader)
	}
}

func TestClientGetConfig(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"VPN_HOST":"vpn.example.com","TCP_SERVICES_PORT_RANGE":"32760-32767"}`))
	}))
	defer srv.Close()

	client, err := NewClient(providerConfig{Endpoint: srv.URL, Token: "token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	cfg, err := client.GetConfig()
	if err != nil {
		t.Fatalf("get config: %v", err)
	}

	if cfg.VPNHost != "vpn.example.com" {
		t.Fatalf("unexpected vpn host: %q", cfg.VPNHost)
	}
	if cfg.TCPServicesPortRange != "32760-32767" {
		t.Fatalf("unexpected port range: %q", cfg.TCPServicesPortRange)
	}
}

func TestResponseErrorContainsStatus(t *testing.T) {
	t.Parallel()

	resp := httptest.NewRecorder()
	resp.WriteHeader(http.StatusNotFound)
	resp.Body.WriteString("missing")

	err := responseError(resp.Result())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error text: %v", err)
	}
}

