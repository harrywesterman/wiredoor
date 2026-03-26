package main

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceHTTPService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHTTPServiceCreate,
		ReadContext:   resourceHTTPServiceRead,
		UpdateContext: resourceHTTPServiceUpdate,
		DeleteContext: resourceHTTPServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceHTTPServiceImport,
		},
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Computed: true},
			"node_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"path_location": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"backend_host": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"backend_port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  80,
			},
			"backend_proto": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "http",
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
			"require_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"skip_auth_routes": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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

func resourceHTTPServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, err := parseID(d.Get("node_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	payload := httpServicePayload(d)
	svc, err := client.CreateHTTPService(nodeID, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	return setHTTPServiceState(d, svc)
}

func resourceHTTPServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	svc, err := client.GetHTTPService(nodeID, serviceID)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setHTTPServiceState(d, svc)
}

func resourceHTTPServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, serviceID, err := parseCompositeID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	svc, err := client.UpdateHTTPService(nodeID, serviceID, httpServicePayload(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return setHTTPServiceState(d, svc)
}

func resourceHTTPServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, serviceID, err := parseCompositeID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.DeleteHTTPService(nodeID, serviceID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourceHTTPServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	nodeID, serviceID, err := parseCompositeID(d.Id())
	if err != nil {
		return nil, err
	}
	d.SetId(strconv.FormatInt(nodeID, 10) + "/" + strconv.FormatInt(serviceID, 10))
	return []*schema.ResourceData{d}, nil
}

func httpServicePayload(d *schema.ResourceData) map[string]interface{} {
	payload := map[string]interface{}{
		"name":           d.Get("name").(string),
		"domain":         d.Get("domain").(string),
		"pathLocation":   d.Get("path_location").(string),
		"backendHost":    d.Get("backend_host").(string),
		"backendPort":    d.Get("backend_port").(int),
		"backendProto":   d.Get("backend_proto").(string),
		"requireAuth":    d.Get("require_auth").(bool),
		"skipAuthRoutes": d.Get("skip_auth_routes").(string),
		"enabled":        d.Get("enabled").(bool),
		"ttl":            d.Get("ttl").(string),
	}
	if v, ok := d.GetOk("allowed_ips"); ok {
		payload["allowedIps"] = stringSlice(v)
	}
	if v, ok := d.GetOk("blocked_ips"); ok {
		payload["blockedIps"] = stringSlice(v)
	}
	return payload
}

func setHTTPServiceState(d *schema.ResourceData, svc *httpService) diag.Diagnostics {
	d.SetId(strconv.FormatInt(svc.NodeID, 10) + "/" + strconv.FormatInt(svc.ID, 10))
	_ = d.Set("node_id", strconv.FormatInt(svc.NodeID, 10))
	_ = d.Set("name", svc.Name)
	_ = d.Set("domain", svc.Domain)
	_ = d.Set("path_location", svc.PathLocation)
	_ = d.Set("backend_host", svc.BackendHost)
	_ = d.Set("backend_port", svc.BackendPort)
	_ = d.Set("backend_proto", svc.BackendProto)
	_ = d.Set("allowed_ips", stringSliceToInterface(svc.AllowedIPs))
	_ = d.Set("blocked_ips", stringSliceToInterface(svc.BlockedIPs))
	_ = d.Set("require_auth", svc.RequireAuth)
	_ = d.Set("skip_auth_routes", svc.SkipAuthRoutes)
	_ = d.Set("enabled", svc.Enabled)
	_ = d.Set("ttl", svc.TTL)
	_ = d.Set("public_access", svc.PublicAccess)
	_ = d.Set("expires_at", formatTimePtr(svc.ExpiresAt))
	_ = d.Set("created_at", formatTime(svc.CreatedAt))
	_ = d.Set("updated_at", formatTime(svc.UpdatedAt))
	return nil
}

func dataSourceHTTPService() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceHTTPServiceRead,
		Schema: map[string]*schema.Schema{
			"node_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name":             {Type: schema.TypeString, Computed: true},
			"domain":           {Type: schema.TypeString, Computed: true},
			"path_location":    {Type: schema.TypeString, Computed: true},
			"backend_host":     {Type: schema.TypeString, Computed: true},
			"backend_port":     {Type: schema.TypeInt, Computed: true},
			"backend_proto":    {Type: schema.TypeString, Computed: true},
			"allowed_ips":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"blocked_ips":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"require_auth":     {Type: schema.TypeBool, Computed: true},
			"skip_auth_routes": {Type: schema.TypeString, Computed: true},
			"enabled":          {Type: schema.TypeBool, Computed: true},
			"ttl":              {Type: schema.TypeString, Computed: true},
			"public_access":    {Type: schema.TypeString, Computed: true},
			"expires_at":       {Type: schema.TypeString, Computed: true},
			"created_at":       {Type: schema.TypeString, Computed: true},
			"updated_at":       {Type: schema.TypeString, Computed: true},
		},
	}
}
