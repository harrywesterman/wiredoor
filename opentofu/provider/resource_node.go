package main

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNode() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNodeCreate,
		ReadContext:   resourceNodeRead,
		UpdateContext: resourceNodeUpdate,
		DeleteContext: resourceNodeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dns": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"keepalive": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  25,
			},
			"address": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"allow_internet": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"advanced": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"is_gateway": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"gateway_networks": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interface": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"gateway_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wg_interface": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"is_local": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNodeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	n, err := client.CreateNode(nodePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return setNodeState(d, n)
}

func resourceNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	idValue := d.Id()
	if idValue == "" {
		idValue = d.Get("id").(string)
	}
	id, err := parseID(idValue)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	n, err := client.GetNode(id)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setNodeState(d, n)
}

func resourceNodeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	n, err := client.UpdateNode(id, nodePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return setNodeState(d, n)
}

func resourceNodeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.DeleteNode(id); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func setNodeState(d *schema.ResourceData, n *node) diag.Diagnostics {
	d.SetId(strconv.FormatInt(n.ID, 10))
	_ = d.Set("name", n.Name)
	_ = d.Set("dns", n.DNS)
	_ = d.Set("keepalive", n.Keepalive)
	_ = d.Set("address", n.Address)
	_ = d.Set("allow_internet", n.AllowInternet)
	_ = d.Set("advanced", n.Advanced)
	_ = d.Set("enabled", n.Enabled)
	_ = d.Set("is_gateway", n.IsGateway)
	_ = d.Set("gateway_network", n.GatewayNetwork)
	_ = d.Set("wg_interface", n.WGInterface)
	_ = d.Set("is_local", n.IsLocal)
	_ = d.Set("created_at", formatTime(n.CreatedAt))
	_ = d.Set("updated_at", formatTime(n.UpdatedAt))
	if len(n.GatewayNetworks) > 0 {
		_ = d.Set("gateway_networks", gatewayNetworksFromAPI(n.GatewayNetworks))
	}
	if n.Token != "" {
		_ = d.Set("token", n.Token)
	}
	return nil
}

func nodePayload(d *schema.ResourceData) map[string]interface{} {
	payload := map[string]interface{}{
		"name":          d.Get("name").(string),
		"dns":           d.Get("dns").(string),
		"keepalive":     d.Get("keepalive").(int),
		"address":       d.Get("address").(string),
		"allowInternet": d.Get("allow_internet").(bool),
		"advanced":      d.Get("advanced").(bool),
		"enabled":       d.Get("enabled").(bool),
		"isGateway":     d.Get("is_gateway").(bool),
	}
	if v, ok := d.GetOk("gateway_networks"); ok {
		payload["gatewayNetworks"] = gatewayNetworksToAPI(v)
	}
	return payload
}

func dataSourceNode() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceNodeRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"keepalive": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_internet": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"advanced": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_gateway": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gateway_networks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interface": {Type: schema.TypeString, Computed: true},
						"subnet":    {Type: schema.TypeString, Computed: true},
					},
				},
			},
			"gateway_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wg_interface": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_local": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
