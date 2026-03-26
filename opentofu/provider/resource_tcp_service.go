package main

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTCPService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTCPServiceCreate,
		ReadContext:   resourceTCPServiceRead,
		UpdateContext: resourceTCPServiceUpdate,
		DeleteContext: resourceTCPServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTCPServiceImport,
		},
		Schema: map[string]*schema.Schema{
			"id":      {Type: schema.TypeString, Computed: true},
			"node_id": {Type: schema.TypeString, Required: true},
			"name":    {Type: schema.TypeString, Required: true},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"proto": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "tcp",
			},
			"backend_host": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"backend_port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allowed_ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"blocked_ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ttl": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"public_access": {Type: schema.TypeString, Computed: true},
			"expires_at":    {Type: schema.TypeString, Computed: true},
			"created_at":    {Type: schema.TypeString, Computed: true},
			"updated_at":    {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceTCPServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, err := parseID(d.Get("node_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	svc, err := client.CreateTCPService(nodeID, tcpServicePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return setTCPServiceState(d, svc)
}

func resourceTCPServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	compositeID := d.Id()
	if compositeID == "" {
		compositeID = d.Get("node_id").(string) + "/" + d.Get("id").(string)
	}
	nodeID, serviceID, err := parseCompositeID(compositeID)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	svc, err := client.GetTCPService(nodeID, serviceID)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setTCPServiceState(d, svc)
}

func resourceTCPServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, serviceID, err := parseCompositeID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	svc, err := client.UpdateTCPService(nodeID, serviceID, tcpServicePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return setTCPServiceState(d, svc)
}

func resourceTCPServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, serviceID, err := parseCompositeID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.DeleteTCPService(nodeID, serviceID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceTCPServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	serviceID, err := parseID(d.Id())
	if err != nil {
		return nil, err
	}
	nodeID := int64(0)
	if raw, ok := d.GetOk("node_id"); ok {
		nodeID = int64Value(raw)
	}
	d.SetId(strconv.FormatInt(nodeID, 10) + "/" + strconv.FormatInt(serviceID, 10))
	return []*schema.ResourceData{d}, nil
}

func tcpServicePayload(d *schema.ResourceData) map[string]interface{} {
	payload := map[string]interface{}{
		"name":         d.Get("name").(string),
		"domain":       d.Get("domain").(string),
		"proto":        d.Get("proto").(string),
		"backendHost":  d.Get("backend_host").(string),
		"backendPort":  d.Get("backend_port").(int),
		"port":         d.Get("port").(int),
		"ssl":          d.Get("ssl").(bool),
		"enabled":      d.Get("enabled").(bool),
		"ttl":          d.Get("ttl").(string),
	}
	if v, ok := d.GetOk("allowed_ips"); ok {
		payload["allowedIps"] = stringSlice(v)
	}
	if v, ok := d.GetOk("blocked_ips"); ok {
		payload["blockedIps"] = stringSlice(v)
	}
	return payload
}

func setTCPServiceState(d *schema.ResourceData, svc *tcpService) diag.Diagnostics {
	d.SetId(strconv.FormatInt(svc.NodeID, 10) + "/" + strconv.FormatInt(svc.ID, 10))
	_ = d.Set("node_id", strconv.FormatInt(svc.NodeID, 10))
	_ = d.Set("name", svc.Name)
	_ = d.Set("domain", svc.Domain)
	_ = d.Set("proto", svc.Proto)
	_ = d.Set("backend_host", svc.BackendHost)
	_ = d.Set("backend_port", svc.BackendPort)
	_ = d.Set("port", svc.Port)
	_ = d.Set("ssl", svc.SSL)
	_ = d.Set("allowed_ips", stringSliceToInterface(svc.AllowedIPs))
	_ = d.Set("blocked_ips", stringSliceToInterface(svc.BlockedIPs))
	_ = d.Set("enabled", svc.Enabled)
	_ = d.Set("ttl", svc.TTL)
	_ = d.Set("public_access", svc.PublicAccess)
	_ = d.Set("expires_at", formatTimePtr(svc.ExpiresAt))
	_ = d.Set("created_at", formatTime(svc.CreatedAt))
	_ = d.Set("updated_at", formatTime(svc.UpdatedAt))
	return nil
}

func dataSourceTCPService() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceTCPServiceRead,
		Schema: map[string]*schema.Schema{
			"node_id": {Type: schema.TypeString, Required: true},
			"id":      {Type: schema.TypeString, Required: true},
			"name":    {Type: schema.TypeString, Computed: true},
			"domain":  {Type: schema.TypeString, Computed: true},
			"proto":   {Type: schema.TypeString, Computed: true},
			"backend_host": {Type: schema.TypeString, Computed: true},
			"backend_port": {Type: schema.TypeInt, Computed: true},
			"port":    {Type: schema.TypeInt, Computed: true},
			"ssl":     {Type: schema.TypeBool, Computed: true},
			"allowed_ips": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"blocked_ips": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"enabled": {Type: schema.TypeBool, Computed: true},
			"ttl":     {Type: schema.TypeString, Computed: true},
			"public_access": {Type: schema.TypeString, Computed: true},
			"expires_at":    {Type: schema.TypeString, Computed: true},
			"created_at":    {Type: schema.TypeString, Computed: true},
			"updated_at":    {Type: schema.TypeString, Computed: true},
		},
	}
}
