package main

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		UpdateContext: resourceDomainUpdate,
		DeleteContext: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Computed: true},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ssl": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "self-signed",
			},
			"authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allowed_emails": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"skip_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"created_at": {Type: schema.TypeString, Computed: true},
			"updated_at": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	domain, err := client.CreateDomain(domainPayload(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return setDomainState(d, domain)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	domain, err := client.GetDomain(id)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return setDomainState(d, domain)
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	domain, err := client.UpdateDomain(id, domainPayload(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return setDomainState(d, domain)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	id, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.DeleteDomain(id); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func domainPayload(d *schema.ResourceData) map[string]interface{} {
	payload := map[string]interface{}{
		"domain":         d.Get("domain").(string),
		"ssl":            d.Get("ssl").(string),
		"authentication": d.Get("authentication").(bool),
		"skipValidation": d.Get("skip_validation").(bool),
	}
	if v, ok := d.GetOk("allowed_emails"); ok {
		payload["allowedEmails"] = stringSlice(v)
	}
	return payload
}

func setDomainState(d *schema.ResourceData, dom *domain) diag.Diagnostics {
	d.SetId(strconv.FormatInt(dom.ID, 10))
	_ = d.Set("domain", dom.Domain)
	_ = d.Set("ssl", dom.SSL)
	_ = d.Set("authentication", dom.Authentication)
	_ = d.Set("allowed_emails", stringSliceToInterface(dom.AllowedEmails))
	_ = d.Set("skip_validation", dom.SkipValidation)
	_ = d.Set("created_at", formatTime(dom.CreatedAt))
	_ = d.Set("updated_at", formatTime(dom.UpdatedAt))
	return nil
}

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceDomainRead,
		Schema: map[string]*schema.Schema{
			"id":              {Type: schema.TypeString, Required: true},
			"domain":          {Type: schema.TypeString, Computed: true},
			"ssl":             {Type: schema.TypeString, Computed: true},
			"authentication":  {Type: schema.TypeBool, Computed: true},
			"allowed_emails":  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"skip_validation":  {Type: schema.TypeBool, Computed: true},
			"created_at":      {Type: schema.TypeString, Computed: true},
			"updated_at":      {Type: schema.TypeString, Computed: true},
		},
	}
}
