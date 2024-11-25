package iam

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3.0/security"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceIdentityPasswordPolicyV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityPasswordPolicyV3Create,
		ReadContext:   resourceIdentityPasswordPolicyV3Read,
		UpdateContext: resourceIdentityPasswordPolicyV3Update,
		DeleteContext: resourceIdentityPasswordPolicyV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"maximum_consecutive_identical_chars": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 32),
			},
			"minimum_password_age": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 1440),
			},
			"minimum_password_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      8,
				ValidateFunc: validation.IntBetween(6, 32),
			},
			"number_of_recent_passwords_disallowed": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"password_not_username_or_invert": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"password_validity_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 180),
			},
			"password_char_combination": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(2, 4),
			},
			"maximum_password_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"password_requirements": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIdentityPasswordPolicyV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	passPolicyOpts := security.UpdatePasswordPolicyOpts{
		MaximumConsecutiveIdenticalChars:  pointerto.Int(d.Get("maximum_consecutive_identical_chars").(int)),
		MinimumPasswordAge:                pointerto.Int(d.Get("minimum_password_age").(int)),
		MinimumPasswordLength:             pointerto.Int(d.Get("minimum_password_length").(int)),
		NumberOfRecentPasswordsDisallowed: pointerto.Int(d.Get("number_of_recent_passwords_disallowed").(int)),
		PasswordNotUsernameOrInvert:       pointerto.Bool(d.Get("password_not_username_or_invert").(bool)),
		PasswordValidityPeriod:            pointerto.Int(d.Get("password_validity_period").(int)),
		PasswordCharCombination:           pointerto.Int(d.Get("password_char_combination").(int)),
	}
	_, err = security.UpdatePasswordPolicy(client, domainID, passPolicyOpts)
	if err != nil {
		return diag.Errorf("error updating the IAM account password policy: %s", err)
	}

	d.SetId(domainID)

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceIdentityPasswordPolicyV3Read(clientCtx, d, meta)
}

func resourceIdentityPasswordPolicyV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	passPolicy, err := security.GetPasswordPolicy(client, d.Id())
	if err != nil {
		return diag.Errorf("error fetching the IAM account password policy")
	}

	log.Printf("[DEBUG] Retrieved the IAM account password policy: %#v", passPolicy)

	mErr := multierror.Append(nil,
		d.Set("number_of_recent_passwords_disallowed", passPolicy.NumberOfRecentPasswordsDisallowed),
		d.Set("minimum_password_age", passPolicy.MinimumPasswordAge),
		d.Set("maximum_consecutive_identical_chars", passPolicy.MaximumConsecutiveIdenticalChars),
		d.Set("minimum_password_length", passPolicy.MinimumPasswordLength),
		d.Set("maximum_password_length", passPolicy.MaximumPasswordLength),
		d.Set("password_not_username_or_invert", passPolicy.PasswordNotUsernameOrInvert),
		d.Set("password_validity_period", passPolicy.PasswordValidityPeriod),
		d.Set("password_requirements", passPolicy.PasswordRequirements),
		d.Set("password_char_combination", passPolicy.PasswordCharCombination),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error setting IAM policy fields: %s", err)
	}
	return nil
}

func resourceIdentityPasswordPolicyV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	if d.HasChanges("maximum_consecutive_identical_chars", "minimum_password_age",
		"minimum_password_length", "number_of_recent_passwords_disallowed",
		"password_not_username_or_invert", "password_validity_period", "password_char_combination") {
		passPolicyOpts := security.UpdatePasswordPolicyOpts{
			MaximumConsecutiveIdenticalChars:  pointerto.Int(d.Get("maximum_consecutive_identical_chars").(int)),
			MinimumPasswordAge:                pointerto.Int(d.Get("minimum_password_age").(int)),
			MinimumPasswordLength:             pointerto.Int(d.Get("minimum_password_length").(int)),
			NumberOfRecentPasswordsDisallowed: pointerto.Int(d.Get("number_of_recent_passwords_disallowed").(int)),
			PasswordNotUsernameOrInvert:       pointerto.Bool(d.Get("password_not_username_or_invert").(bool)),
			PasswordValidityPeriod:            pointerto.Int(d.Get("password_validity_period").(int)),
			PasswordCharCombination:           pointerto.Int(d.Get("password_char_combination").(int)),
		}
		_, err = security.UpdatePasswordPolicy(client, d.Id(), passPolicyOpts)
		if err != nil {
			return diag.Errorf("error updating the IAM password policy: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceIdentityPasswordPolicyV3Read(clientCtx, d, meta)
}

func resourceIdentityPasswordPolicyV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	_, err = security.UpdatePasswordPolicy(client, d.Id(),
		security.UpdatePasswordPolicyOpts{
			MaximumConsecutiveIdenticalChars:  pointerto.Int(0),
			MinimumPasswordAge:                pointerto.Int(0),
			MinimumPasswordLength:             pointerto.Int(8),
			NumberOfRecentPasswordsDisallowed: pointerto.Int(1),
			PasswordNotUsernameOrInvert:       pointerto.Bool(true),
			PasswordValidityPeriod:            pointerto.Int(0),
			PasswordCharCombination:           pointerto.Int(2),
		})
	if err != nil {
		return diag.Errorf("error resetting the IAM account password policy: %s", err)
	}

	return nil
}
