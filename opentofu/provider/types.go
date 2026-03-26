package main

import "time"

type providerConfig struct {
	Endpoint string
	Username string
	Password string
	Token    string
	Insecure bool
	CACert   string
}

type authResponse struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expiresIn"`
}

type publicConfig struct {
	VPNHost              string `json:"VPN_HOST"`
	TCPServicesPortRange string `json:"TCP_SERVICES_PORT_RANGE"`
}

type gatewayNetwork struct {
	Interface string `json:"interface"`
	Subnet    string `json:"subnet"`
}

type node struct {
	ID              int64           `json:"id"`
	Name            string          `json:"name"`
	Address         string          `json:"address"`
	DNS             string          `json:"dns"`
	Keepalive       int64           `json:"keepalive"`
	GatewayNetwork  string          `json:"gatewayNetwork"`
	GatewayNetworks []gatewayNetwork `json:"gatewayNetworks"`
	WGInterface     string          `json:"wgInterface"`
	AllowInternet   bool            `json:"allowInternet"`
	Advanced        bool            `json:"advanced"`
	Enabled         bool            `json:"enabled"`
	IsGateway       bool            `json:"isGateway"`
	IsLocal         bool            `json:"isLocal"`
	Token           string          `json:"token,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type httpService struct {
	ID           int64    `json:"id"`
	NodeID       int64    `json:"nodeId"`
	Name         string   `json:"name"`
	Domain       string   `json:"domain"`
	PathLocation string   `json:"pathLocation"`
	BackendHost  string   `json:"backendHost"`
	BackendPort  int64    `json:"backendPort"`
	BackendProto string   `json:"backendProto"`
	AllowedIPs   []string `json:"allowedIps"`
	BlockedIPs   []string `json:"blockedIps"`
	RequireAuth  bool     `json:"requireAuth"`
	SkipAuthRoutes string `json:"skipAuthRoutes"`
	Enabled      bool     `json:"enabled"`
	TTL          string   `json:"ttl"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	PublicAccess string   `json:"publicAccess"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type tcpService struct {
	ID          int64      `json:"id"`
	NodeID      int64      `json:"nodeId"`
	Name        string     `json:"name"`
	Domain      string     `json:"domain"`
	Proto       string     `json:"proto"`
	BackendHost string     `json:"backendHost"`
	BackendPort int64      `json:"backendPort"`
	Port        int64      `json:"port"`
	SSL         bool       `json:"ssl"`
	AllowedIPs  []string   `json:"allowedIps"`
	BlockedIPs  []string   `json:"blockedIps"`
	Enabled     bool       `json:"enabled"`
	TTL         string     `json:"ttl"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	PublicAccess string    `json:"publicAccess"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type domain struct {
	ID             int64      `json:"id"`
	Domain         string     `json:"domain"`
	SSL            string     `json:"ssl"`
	Authentication bool       `json:"authentication"`
	AllowedEmails  []string   `json:"allowedEmails"`
	SkipValidation bool       `json:"skipValidation"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type pat struct {
	ID        int64      `json:"id"`
	NodeID    int64      `json:"nodeId"`
	Name      string     `json:"name"`
	ExpireAt  *time.Time `json:"expireAt"`
	Revoked   bool       `json:"revoked"`
	Token     string     `json:"token,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func gatewayNetworksToAPI(v interface{}) []gatewayNetwork {
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]gatewayNetwork, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		result = append(result, gatewayNetwork{
			Interface: stringValue(m["interface"]),
			Subnet:    stringValue(m["subnet"]),
		})
	}
	return result
}

func gatewayNetworksFromAPI(items []gatewayNetwork) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]interface{}{
			"interface": item.Interface,
			"subnet":    item.Subnet,
		})
	}
	return result
}

func stringSlice(v interface{}) []string {
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func stringSliceToInterface(items []string) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func stringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

