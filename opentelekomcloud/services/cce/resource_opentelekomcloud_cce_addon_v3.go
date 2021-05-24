package cce

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/cce/v3/addons"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

func ResourceCCEAddonV3() *schema.Resource {
	return &schema.Resource{
		Create: resourceCCEAddonV3Create,
		Read:   resourceCCEAddonV3Read,
		Delete: resourceCCEAddonV3Delete,

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: resourceCCEAddonV3Import,
		},

		Schema: map[string]*schema.Schema{
			"template_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"template_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"values": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"basic": {
							Type:     schema.TypeMap,
							Required: true,
							ForceNew: true,
						},
						"custom": {
							Type:     schema.TypeMap,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceCCEAddonV3Create(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating CCE client: %w", err)
	}

	clusterID := d.Get("cluster_id").(string)
	basic, custom, err := getAddonValues(d)
	if err != nil {
		return fmt.Errorf("error getting values for CCE addon: %w", err)
	}

	basic = unStringMap(basic)
	custom = unStringMap(custom)

	templateName := d.Get("template_name").(string)
	addon, err := addons.Create(client, addons.CreateOpts{
		Kind:       "Addon",
		ApiVersion: "v3",
		Metadata: addons.CreateMetadata{
			Annotations: addons.CreateAnnotations{
				AddonInstallType: "install",
			},
		},
		Spec: addons.RequestSpec{
			Version:           d.Get("template_version").(string),
			ClusterID:         clusterID,
			AddonTemplateName: templateName,
			Values: addons.Values{
				Basic:    basic,
				Advanced: custom,
			},
		},
	}, clusterID).Extract()

	if err != nil {
		errMsg := logHttpError(err)
		addonSpec, aErr := getAddonTemplateSpec(client, clusterID, templateName)
		if aErr == nil {
			errMsg = fmt.Errorf("\nAddon template spec: %s\n%s", addonSpec, errMsg)
		}
		return fmt.Errorf("error creating CCE addon instance: %w", errMsg)
	}

	d.SetId(addon.Metadata.Id)

	return resourceCCEAddonV3Read(d, meta)
}

func resourceCCEAddonV3Read(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating CCE client: %w", logHttpError(err))
	}

	clusterID := d.Get("cluster_id").(string)
	addon, err := addons.Get(client, d.Id(), clusterID).Extract()
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading CCE addon instance: %w", logHttpError(err))
	}

	mErr := multierror.Append(nil,
		d.Set("name", addon.Metadata.Name),
		d.Set("cluster_id", addon.Spec.ClusterID),
		d.Set("template_version", addon.Spec.Version),
		d.Set("template_name", addon.Spec.AddonTemplateName),
		d.Set("description", addon.Spec.Description),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return fmt.Errorf("error setting addon attributes: %w", err)
	}

	return nil
}

func getAddonValues(d *schema.ResourceData) (basic, custom map[string]interface{}, err error) {
	valLength := d.Get("values.#").(int)
	if valLength == 0 {
		err = fmt.Errorf("no values are set for CCE addon")
		return
	}
	basic = d.Get("values.0.basic").(map[string]interface{})
	custom = d.Get("values.0.custom").(map[string]interface{})
	return
}

func resourceCCEAddonV3Delete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*cfg.Config)
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return fmt.Errorf("error creating CCE client: %w", err)
	}
	clusterID := d.Get("cluster_id").(string)

	stateConf := &resource.StateChangeConf{
		Pending: []string{"available"},
		Target:  []string{"deleted"},
		Refresh: waitForCCEAddonDelete(client, d.Id(), clusterID),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func getAddonTemplateSpec(client *golangsdk.ServiceClient, clusterID, templateName string) (string, error) {
	templates, err := addons.ListTemplates(client, clusterID, addons.ListOpts{Name: templateName}).Extract()
	if err != nil {
		return "", err
	}
	jsonTemplate, _ := json.Marshal(templates)
	return string(jsonTemplate), nil
}

func logHttpError(err error) error {
	if httpErr, ok := err.(golangsdk.ErrUnexpectedResponseCode); ok {
		return fmt.Errorf("response: %s\n %s", httpErr.Error(), httpErr.Body)
	}
	return err
}

// Make map values to be not a string, if possible
func unStringMap(src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(src))
	for key, v := range src {
		val := v.(string)
		if intVal, err := strconv.Atoi(val); err == nil {
			out[key] = intVal
			continue
		}
		if boolVal, err := strconv.ParseBool(val); err == nil {
			out[key] = boolVal
			continue
		}
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			out[key] = floatVal
			continue
		}
		out[key] = val
	}
	return out
}

func waitForCCEAddonDelete(client *golangsdk.ServiceClient, addonID, clusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if err := addons.Delete(client, addonID, clusterID).ExtractErr(); err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return nil, "deleted", nil
			}
			return nil, "error", fmt.Errorf("error deleting CCE addon : %w", err)
		}

		addon, err := addons.Get(client, addonID, clusterID).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return addon, "deleted", nil
			}
			return nil, "error", fmt.Errorf("error waiting CCE addon to become deleted: %w", err)
		}

		return addon, "available", nil
	}
}

func resourceCCEAddonV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("invalid format specified for CCE Addon. Format must be <cluster id>/<addon id>")
		return nil, err
	}
	clusterID := parts[0]
	addonID := parts[1]
	d.SetId(addonID)

	config := meta.(*cfg.Config)
	client, err := config.CceV3AddonClient(config.GetRegion(d))
	if err != nil {
		return nil, fmt.Errorf("error creating CCE client: %w", logHttpError(err))
	}

	addon, err := addons.Get(client, d.Id(), clusterID).Extract()
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil, fmt.Errorf("addon wasn't found")
		}

		return nil, fmt.Errorf("error reading CCE addon instance: %w", logHttpError(err))
	}

	mErr := multierror.Append(nil,
		d.Set("name", addon.Metadata.Name),
		d.Set("cluster_id", addon.Spec.ClusterID),
		d.Set("template_version", addon.Spec.Version),
		d.Set("template_name", addon.Spec.AddonTemplateName),
		d.Set("description", addon.Spec.Description),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return nil, fmt.Errorf("error setting addon attributes: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}
