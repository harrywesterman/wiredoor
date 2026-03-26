package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigRead,
		Schema: map[string]*schema.Schema{
			"vpn_host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tcp_services_port_range": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	cfg, err := client.GetConfig()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("config")
	_ = d.Set("vpn_host", cfg.VPNHost)
	_ = d.Set("tcp_services_port_range", cfg.TCPServicesPortRange)
	return nil
}
