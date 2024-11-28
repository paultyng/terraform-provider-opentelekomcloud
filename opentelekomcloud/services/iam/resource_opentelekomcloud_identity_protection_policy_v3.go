package iam

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3.0/security"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceIdentityProtectionPolicyV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityProtectionPolicyV3Create,
		ReadContext:   resourceIdentityProtectionPolicyV3Read,
		UpdateContext: resourceIdentityProtectionPolicyV3Create,
		DeleteContext: resourceIdentityProtectionPolicyV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enable_operation_protection_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"verification_mobile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"verification_email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"self_management": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"password": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"mobile": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"email": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"self_verification": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceIdentityProtectionPolicyV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	domainID, err := getDomainID(config, client)
	if err != nil {
		return fmterr.Errorf("error getting the domain id, err=%s", err)
	}

	enable := d.Get("enable_operation_protection_policy").(bool)
	opPolicyOpts := security.UpdateProtectionPolicyOpts{
		OperationProtection: pointerto.Bool(enable),
		AllowUser:           buildSelfManagement(d),
	}

	// verification_mobile and verification_mobile are valid when the protection is enabled
	if enable {
		var adminCheck string
		if v, ok := d.GetOk("verification_mobile"); ok {
			adminCheck = "on"
			opPolicyOpts.Scene = "mobile"
			opPolicyOpts.Mobile = v.(string)
		} else if v, ok := d.GetOk("verification_email"); ok {
			adminCheck = "on"
			opPolicyOpts.Scene = "email"
			opPolicyOpts.Email = v.(string)
		} else {
			// self verification
			adminCheck = "off"
		}
		opPolicyOpts.AdminCheck = adminCheck
	}

	_, err = security.UpdateOperationProtectionPolicy(client, domainID, opPolicyOpts)
	if err != nil {
		return diag.Errorf("error updating the IAM operation protection policy: %s", err)
	}

	// set the ID only when creating
	if d.IsNewResource() {
		d.SetId(domainID)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceIdentityProtectionPolicyV3Read(clientCtx, d, meta)
}

func resourceIdentityProtectionPolicyV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	opPolicy, err := security.GetOperationProtectionPolicy(client, d.Id())
	if err != nil {
		return diag.Errorf("error fetching the IAM operation protection policy")
	}

	log.Printf("[DEBUG] Retrieved the IAM operation protection policy: %#v", opPolicy)

	mErr := multierror.Append(nil,
		d.Set("enable_operation_protection_policy", opPolicy.OperationProtection),
		d.Set("verification_email", opPolicy.Email),
		d.Set("verification_mobile", opPolicy.Mobile),
		d.Set("self_verification", opPolicy.AdminCheck != "on"),
		d.Set("self_management", flattenSelfManagement(opPolicy.AllowUser)),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error setting IAM policy fields: %s", err)
	}
	return nil
}

func flattenSelfManagement(resp *security.AllowUser) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"access_key": resp.ManageAccessKey,
			"password":   resp.ManagePassword,
			"mobile":     resp.ManageMobile,
			"email":      resp.ManageEmail,
		},
	}
}

func resourceIdentityProtectionPolicyV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	_, err = security.UpdateOperationProtectionPolicy(client, d.Id(),
		security.UpdateProtectionPolicyOpts{OperationProtection: pointerto.Bool(false)},
	)
	if err != nil {
		return diag.Errorf("error resetting the IAM protection policy: %s", err)
	}

	return nil
}

func buildSelfManagement(d *schema.ResourceData) *security.AllowUser {
	raw := d.Get("self_management").([]interface{})
	if len(raw) == 0 {
		// if not specified, keep the previous settings.
		return nil
	}

	item, ok := raw[0].(map[string]interface{})
	if !ok {
		return nil
	}

	allowed := security.AllowUser{}
	if v, ok := item["access_key"]; ok {
		allowed.ManageAccessKey = pointerto.Bool(v.(bool))
	}
	if v, ok := item["password"]; ok {
		allowed.ManagePassword = pointerto.Bool(v.(bool))
	}
	if v, ok := item["mobile"]; ok {
		allowed.ManageMobile = pointerto.Bool(v.(bool))
	}
	if v, ok := item["email"]; ok {
		allowed.ManageEmail = pointerto.Bool(v.(bool))
	}

	return &allowed
}
