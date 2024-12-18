package cce

import (
	"context"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/addons"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/helper/hashcode"
)

func DataSourceCceAddonTemplatesV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCceAddonTemplatesV3Read,

		Schema: map[string]*schema.Schema{
			"cluster_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"addon_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"VirtualMachine", "ARM64", "BareMetal",
				}, false),
				Default: "VirtualMachine",
			},
			"addons": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"addon_version": {
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
						"platform": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"euleros_version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"obs_url": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCceAddonTemplatesV3Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	cVersion := d.Get("cluster_version").(string)
	cType := d.Get("cluster_type").(string)
	if len(cVersion) > 0 && cVersion[0] != 'v' {
		cVersion = "v" + cVersion
	}
	var templates []addons.Version
	var ids []string
	for _, v := range template.Spec.Versions {
		for _, supported := range v.SupportVersions {
			vList := unpackVersions(supported.ClusterVersion)
			for _, current := range vList {
				if matchVersion(cVersion, current) && cType == supported.ClusterType {
					templates = append(templates, v)
					ids = append(ids, v.UpdateTimestamp)
					break
				}
			}
		}
	}
	if len(templates) == 0 {
		return fmterr.Errorf("your query returned no results by provided version." +
			" Please change your search criteria and try again")
	}

	d.SetId(hashcode.Strings(ids))

	result := make([]map[string]interface{}, len(templates))
	for i, item := range templates {
		inputData := item.Input.Basic
		addon := map[string]interface{}{
			"addon_version":   item.Version,
			"cluster_ip":      inputData["cluster_ip"],
			"image_version":   inputData["image_version"],
			"platform":        inputData["platform"],
			"swr_addr":        inputData["swr_addr"],
			"swr_user":        inputData["swr_user"],
			"euleros_version": inputData["euleros_version"],
			"obs_url":         inputData["obs_url"],
		}

		result[i] = addon
	}

	if err := d.Set("addons", result); err != nil {
		return diag.Errorf("error setting CCE addons templates list: %s", err)
	}

	return nil
}

func unpackVersions(version []string) []string {
	var result []string

	for _, str := range version {
		re := regexp.MustCompile(`\((.*?)\)`) // Find all patterns enclosed in round brackets
		matches := re.FindAllStringSubmatch(str, -1)
		if len(matches) > 0 {
			options := strings.Split(matches[0][1], "|") // Split the options separated by |
			for _, option := range options {
				result = append(result, strings.Replace(str, matches[0][0], option, 1)) // Replace the matched pattern with each option
			}
		} else {
			result = append(result, str)
		}
	}

	return result
}

func matchVersion(version, pattern string) bool {
	match, _ := regexp.MatchString("^"+pattern+"$", version)
	if match {
		return match
	}
	// Handle wildcard pattern
	wildcardPattern := pattern[:len(pattern)-3]
	match, _ = regexp.MatchString("^"+wildcardPattern+"$", version)
	return match
}
