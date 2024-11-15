package kms

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceKmsImportParamsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKmsImportParamsV1Read,

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"wrapping_algorithm": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"RSAES_PKCS1_V1_5",
					"RSAES_OAEP_SHA_1",
					"RSAES_OAEP_SHA_256",
				}, false),
			},
			"sequence": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"import_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKmsImportParamsV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	client, err := config.KmsKeyV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud kms key client: %s", err)
	}

	opts := keys.GetCMKImportOpts{
		KeyId:             d.Get("key_id").(string),
		WrappingAlgorithm: d.Get("wrapping_algorithm").(string),
		Sequence:          d.Get("sequence").(string),
	}

	params, err := keys.GetCMKImport(client, opts)
	if err != nil {
		if strings.Contains(err.Error(), "The key is not pending for import") {
			keyGet, err := keys.Get(client, d.Get("key_id").(string))
			if err != nil {
				return fmterr.Errorf("error getting key state: %s", err)
			}

			switch keyGet.KeyState {
			case "4":
				d.SetId("")
				return nil
			case "2":
				d.SetId(d.Get("key_id").(string))
				mErr := multierror.Append(
					d.Set("import_token", d.Get("import_token").(string)),
					d.Set("expiration_time", d.Get("expiration_time").(int)),
					d.Set("public_key", d.Get("public_key").(string)),
				)
				if err := mErr.ErrorOrNil(); err != nil {
					return diag.FromErr(err)
				}
				return nil
			default:
				return fmterr.Errorf("unexpected key state: %s", keyGet.KeyState)
			}
		}
		return fmterr.Errorf("error getting CMK import parameters: %s", err)
	}

	d.SetId(d.Get("key_id").(string))
	mErr := multierror.Append(
		d.Set("import_token", params.ImportToken),
		d.Set("expiration_time", params.ExpirationTime),
		d.Set("public_key", params.PublicKey),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
