package apigw

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/apigw/v2/cert"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceAPICertificateV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCertificateV2Create,
		ReadContext:   resourceCertificateV2Read,
		UpdateContext: resourceCertificateV2Update,
		DeleteContext: resourceCertificateV2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"content": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"trusted_root_ca": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"effected_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signature_algorithm": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sans": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func buildCreateCertificateOpts(d *schema.ResourceData) cert.CreateOpts {
	opts := cert.CreateOpts{
		Name:          d.Get("name").(string),
		CertContent:   d.Get("content").(string),
		PrivateKey:    d.Get("private_key").(string),
		Type:          d.Get("type").(string),
		InstanceID:    d.Get("instance_id").(string),
		TrustedRootCA: d.Get("trusted_root_ca").(string),
	}
	return opts
}

func buildUpdateCertificateOpts(d *schema.ResourceData) cert.UpdateOpts {
	opts := cert.UpdateOpts{
		Name:          d.Get("name").(string),
		CertContent:   d.Get("content").(string),
		PrivateKey:    d.Get("private_key").(string),
		Type:          d.Get("type").(string),
		InstanceID:    d.Get("instance_id").(string),
		TrustedRootCA: d.Get("trusted_root_ca").(string),
	}
	return opts
}

func resourceCertificateV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	resp, err := cert.Create(client, buildCreateCertificateOpts(d))
	if err != nil {
		return diag.Errorf("error creating APIG SSL certificate: %s", err)
	}
	d.SetId(resp.ID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceCertificateV2Read(clientCtx, d, meta)
}

func resourceCertificateV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	certificateId := d.Id()
	resp, err := cert.Get(client, certificateId)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "APIG SSL certificate")
	}

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("name", resp.Name),
		d.Set("type", resp.Type),
		d.Set("instance_id", resp.InstanceID),
		d.Set("effected_at", resp.NotBefore),
		d.Set("expires_at", resp.NotAfter),
		d.Set("signature_algorithm", resp.SignatureAlgorithm),
		d.Set("sans", resp.San),
	)
	if mErr.ErrorOrNil() != nil {
		return diag.Errorf("error saving APIG SSL certificate (%s) fields: %s", certificateId, mErr)
	}
	return nil
}

func resourceCertificateV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	_, err = cert.Update(client, d.Id(), buildUpdateCertificateOpts(d))
	if err != nil {
		return diag.Errorf("error updating APIG SSL certificate: %s", err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV2)
	return resourceCertificateV2Read(clientCtx, d, meta)
}

func resourceCertificateV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.APIGWV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV2Client, err)
	}

	certificateId := d.Id()
	err = cert.Delete(client, certificateId)
	if err != nil {
		return common.CheckDeletedDiag(d, err, fmt.Sprintf("error deleting APIG SSL certificate (%s): %s",
			certificateId, err))
	}

	return nil
}
