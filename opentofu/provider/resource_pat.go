package main

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePAT() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePATCreate,
		ReadContext:   resourcePATRead,
		DeleteContext: resourcePATDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePATImport,
		},
		Schema: map[string]*schema.Schema{
			"id":         {Type: schema.TypeString, Computed: true},
			"node_id":    {Type: schema.TypeString, Required: true, ForceNew: true},
			"name":       {Type: schema.TypeString, Required: true, ForceNew: true},
			"token":      {Type: schema.TypeString, Computed: true, Sensitive: true},
			"revoked":    {Type: schema.TypeBool, Computed: true},
			"created_at": {Type: schema.TypeString, Computed: true},
			"updated_at": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourcePATCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, err := parseID(d.Get("node_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := client.CreatePAT(nodeID, map[string]interface{}{
		"name": d.Get("name").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return setPATState(d, p)
}

func resourcePATRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, patID, err := parseCompositeID(d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	p, err := client.GetPAT(nodeID, patID)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setPATState(d, p)
}

func resourcePATDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	nodeID, patID, err := parseCompositeID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.DeletePAT(nodeID, patID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func resourcePATImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	nodeID, patID, err := parseCompositeID(d.Id())
	if err != nil {
		return nil, err
	}
	d.SetId(strconv.FormatInt(nodeID, 10) + "/" + strconv.FormatInt(patID, 10))
	return []*schema.ResourceData{d}, nil
}

func setPATState(d *schema.ResourceData, p *pat) diag.Diagnostics {
	d.SetId(strconv.FormatInt(p.NodeID, 10) + "/" + strconv.FormatInt(p.ID, 10))
	_ = d.Set("node_id", strconv.FormatInt(p.NodeID, 10))
	_ = d.Set("name", p.Name)
	if p.Token != "" {
		_ = d.Set("token", p.Token)
	} else {
		_ = d.Set("token", "")
	}
	_ = d.Set("revoked", p.Revoked)
	_ = d.Set("created_at", formatTime(p.CreatedAt))
	_ = d.Set("updated_at", formatTime(p.UpdatedAt))
	return nil
}
