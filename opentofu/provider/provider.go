package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Wiredoor server base URL, for example https://wiredoor.example.com",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Admin username for /auth/login.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Admin password for /auth/login.",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Pre-authenticated admin or PAT bearer token.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip TLS certificate verification.",
			},
			"ca_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Optional PEM-encoded CA certificate. Reserved for future use.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"wiredoor_node":        resourceNode(),
			"wiredoor_http_service": resourceHTTPService(),
			"wiredoor_tcp_service":  resourceTCPService(),
			"wiredoor_domain":      resourceDomain(),
			"wiredoor_node_pat":    resourcePAT(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"wiredoor_config":        dataSourceConfig(),
			"wiredoor_node":          dataSourceNode(),
			"wiredoor_http_service":   dataSourceHTTPService(),
			"wiredoor_tcp_service":   dataSourceTCPService(),
			"wiredoor_domain":        dataSourceDomain(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		cfg := providerConfig{
			Endpoint: d.Get("endpoint").(string),
			Username: d.Get("username").(string),
			Password: d.Get("password").(string),
			Token:    d.Get("token").(string),
			Insecure: d.Get("insecure").(bool),
			CACert:   d.Get("ca_cert").(string),
		}

		if cfg.Token == "" && (cfg.Username == "" || cfg.Password == "") {
			return nil, diag.FromErr(fmt.Errorf("either token or username/password must be configured"))
		}

		client, err := NewClient(cfg)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		if cfg.Token == "" {
			if err := client.Authenticate(cfg.Username, cfg.Password); err != nil {
				return nil, diag.FromErr(err)
			}
		}

		return client, nil
	}

	return p
}
