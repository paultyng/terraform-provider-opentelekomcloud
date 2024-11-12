package kms

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func DataSourceKmsDataKeyV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKmsDataKeyV1Read,

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encryption_context": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"datakey_length": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"plain_text": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cipher_text": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKmsDataKeyV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	KmsDataKeyV1Client, err := config.KmsKeyV1Client(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud kms key client: %s", err)
	}

	req := keys.DataEncryptOpts{
		KeyID:             d.Get("key_id").(string),
		EncryptionContext: d.Get("encryption_context").(string),
		DatakeyLength:     d.Get("datakey_length").(string),
	}
	log.Printf("[DEBUG] KMS get data key for key: %s", d.Get("key_id").(string))
	v, err := keys.DataEncryptGet(KmsDataKeyV1Client, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(time.Now().UTC().String())
	mErr := multierror.Append(
		d.Set("plain_text", v.PlainText),
		d.Set("cipher_text", v.CipherText),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
