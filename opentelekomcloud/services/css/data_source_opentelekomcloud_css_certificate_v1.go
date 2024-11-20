package css

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	csscert "github.com/opentelekomcloud/gophertelekomcloud/openstack/css/v1/certifcates"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/helper/hashcode"
)

func DataSourceCSSCertificateV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCSSCertificateV1Read,
		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCSSCertificateV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.CssV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating CSS v1 client: %s", err)
	}

	cert, err := csscert.Get(client)
	if err != nil {
		return nil
	}
	d.SetId(hashcode.Strings([]string{*cert}))

	mErr := multierror.Append(
		d.Set("project_id", config.TenantID),
		d.Set("region", config.GetRegion(d)),
		d.Set("certificate", cert),
	)

	if mErr.ErrorOrNil() != nil {
		return diag.FromErr(mErr)
	}
	return nil
}
