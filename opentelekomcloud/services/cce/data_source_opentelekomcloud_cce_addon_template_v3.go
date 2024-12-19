package cce

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/addons"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceCceAddonTemplateV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCceAddonTemplateV3Read,

		Schema: map[string]*schema.Schema{
			"addon_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"addon_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_versions": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"swr_addr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"swr_user": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceCceAddonTemplateV3Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf(cceClientError, err)
	}

	addonTemplates, err := addons.GetTemplates(client)
	if err != nil {
		return fmterr.Errorf("unable to retrieve templates: %s", err)
	}
	aName := d.Get("addon_name")
	var template addons.AddonTemplate
	for _, addon := range addonTemplates.Items {
		if addon.Metadata.Name == aName {
			template = addon
			break
		}
	}
	if template.Metadata.Id == "" {
		return fmterr.Errorf("your query returned no results by provided addon name." +
			" Please change your search criteria and try again")
	}

	aVersion := d.Get("addon_version")
	var result addons.Version
	for _, version := range template.Spec.Versions {
		if version.Version == aVersion {
			result = version
			break
		}
	}
	if result.Version == "" {
		return fmterr.Errorf("your query returned no results by provided version." +
			" Please change your search criteria and try again")
	}

	log.Printf("[DEBUG] Retrieved Template using given filter: %s", template.Metadata.Id)
	d.SetId(template.Metadata.Id)

	inputData := result.Input.Basic
	mErr := multierror.Append(
		d.Set("cluster_ip", inputData["cluster_ip"]),
		d.Set("image_version", inputData["image_version"]),
		d.Set("swr_addr", inputData["swr_addr"]),
		d.Set("swr_user", inputData["swr_user"]),
		d.Set("cluster_versions", result.SupportVersions[0].ClusterVersion[0]),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
