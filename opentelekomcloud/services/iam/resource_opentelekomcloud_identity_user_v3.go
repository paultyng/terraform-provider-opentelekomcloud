package iam

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/common/pointerto"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3.0/security"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3.0/users"
	oldusers "github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/users"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceIdentityUserV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityUserV3Create,
		ReadContext:   resourceIdentityUserV3Read,
		UpdateContext: resourceIdentityUserV3Update,
		DeleteContext: resourceIdentityUserV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: common.SuppressCaseInsensitive,
				ValidateFunc:     common.ValidateEmail,
			},
			"phone": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"country_code"},
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9]{0,32}$"),
					"the phone number must have a maximum of 32 digits"),
			},
			"country_code": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"phone"},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"pwd_reset": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"access_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"default", "programmatic", "console",
				}, false),
			},
			"send_welcome_email": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"email"},
			},
			"login_protection": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"verification_method": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"sms", "email", "vmfa",
							}, false),
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"password_strength": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xuser_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xuser_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIdentityUserV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	domainId, err := getDomainID(config, client)
	if err != nil {
		return fmterr.Errorf("domain name or id missing: %s", err)
	}

	enabled := d.Get("enabled").(bool)
	reset := d.Get("pwd_reset").(bool)
	createOpts := users.CreateOpts{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Email:         d.Get("email").(string),
		Phone:         d.Get("phone").(string),
		AreaCode:      d.Get("country_code").(string),
		AccessMode:    d.Get("access_type").(string),
		Enabled:       &enabled,
		PasswordReset: &reset,
		DomainID:      domainId,
	}

	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	// Add password here, so it wouldn't go in the above log entry
	createOpts.Password = d.Get("password").(string)

	user, err := users.CreateUser(client, createOpts)
	if err != nil {
		return diag.Errorf("error creating IAM user: %s", err)
	}

	d.SetId(user.ID)

	if d.Get("send_welcome_email").(bool) {
		timeout := 6 * time.Second // 3 attempts with 2-second delay
		retryErr := resource.RetryContext(ctx, timeout, func() *resource.RetryError {
			err := oldusers.SendWelcomeEmail(client, d.Id()).ExtractErr()
			if err != nil {
				return resource.RetryableError(fmt.Errorf("error sending a welcome email, retrying: %w", err))
			}
			return nil
		})
		if retryErr != nil {
			return fmterr.Errorf("error sending a welcome email: %w", retryErr)
		}
	}

	if pConfig, ok := d.GetOk("login_protection"); ok {
		configMap := pConfig.([]interface{})
		c := configMap[0].(map[string]interface{})
		protectionOpts := security.LoginProtectionUpdateOpts{
			Enabled:            pointerto.Bool(c["enabled"].(bool)),
			VerificationMethod: c["verification_method"].(string),
		}
		_, err = security.UpdateLoginProtectionConfiguration(client, user.ID, protectionOpts)
		if err != nil {
			return diag.Errorf("error updating protection configuration for IAM user: %s", err)
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceIdentityUserV3Read(clientCtx, d, meta)
}

func resourceIdentityUserV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	user, err := users.GetUser(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "user")
	}

	log.Printf("[DEBUG] Retrieved IAM user: %#v", user)
	mErr := multierror.Append(nil,
		d.Set("enabled", user.Enabled),
		d.Set("name", user.Name),
		d.Set("description", user.Description),
		d.Set("email", user.Email),
		d.Set("phone", normalizePhoneNumber(user.Phone)),
		d.Set("country_code", user.AreaCode),
		d.Set("access_type", user.AccessMode),
		d.Set("password_strength", user.PasswordStrength),
		d.Set("pwd_reset", user.PasswordStatus),
		d.Set("create_time", user.CreateAt),
		d.Set("last_login", user.LastLogin),
		d.Set("domain_id", user.DomainID),
		d.Set("xuser_type", user.XuserType),
		d.Set("xuser_id", user.XuserID),
	)

	userProtectionConfig, _ := getLoginProtection(client, d)
	if userProtectionConfig != nil {
		verMethod := userProtectionConfig.VerificationMethod
		stateVerMethod := d.Get("login_protection.0.verification_method").(string)
		if verMethod != "none" {
			protection := []map[string]interface{}{
				{
					"enabled":             userProtectionConfig.Enabled,
					"verification_method": verMethod,
				},
			}
			mErr = multierror.Append(mErr,
				d.Set("login_protection", protection),
			)
		}
		if verMethod == "none" && stateVerMethod != "" {
			verMethod = stateVerMethod
			protection := []map[string]interface{}{
				{
					"enabled":             userProtectionConfig.Enabled,
					"verification_method": verMethod,
				},
			}
			mErr = multierror.Append(mErr,
				d.Set("login_protection", protection),
			)
		}
	}

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error setting IAM user fields: %s", err)
	}
	return nil
}

func getLoginProtection(client *golangsdk.ServiceClient, d *schema.ResourceData) (*security.LoginProtectionConfig, error) {
	userProtectionConfig, err := security.GetLoginProtectionConfiguration(client, d.Id())
	if err != nil {
		return nil, fmt.Errorf("error obtaining user security config %s", err)
	}
	return userProtectionConfig, nil
}

func resourceIdentityUserV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV30Client()
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}

	var hasChange bool
	var updateOpts users.UpdateOpts
	if d.HasChange("name") {
		updateOpts.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		updateOpts.Description = d.Get("description").(string)
	}

	if d.HasChange("email") {
		updateOpts.Email = d.Get("email").(string)
		hasChange = true
	}

	if d.HasChanges("country_code", "phone") {
		updateOpts.AreaCode = d.Get("country_code").(string)
		updateOpts.Phone = d.Get("phone").(string)
	}

	if d.HasChange("access_type") {
		updateOpts.AccessMode = d.Get("access_type").(string)
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		updateOpts.Enabled = &enabled
	}

	if d.HasChange("pwd_reset") {
		reset := d.Get("pwd_reset").(bool)
		updateOpts.PasswordReset = &reset
	}

	log.Printf("[DEBUG] Update Options: %#v", updateOpts)

	// Add password here so it wouldn't go in the above log entry
	if d.HasChange("password") {
		updateOpts.Password = d.Get("password").(string)
	}

	if d.HasChange("login_protection") {
		configMap := d.Get("login_protection").([]interface{})
		if len(configMap) > 0 {
			c := configMap[0].(map[string]interface{})
			protectionOpts := security.LoginProtectionUpdateOpts{
				Enabled:            pointerto.Bool(c["enabled"].(bool)),
				VerificationMethod: c["verification_method"].(string),
			}
			_, err = security.UpdateLoginProtectionConfiguration(client, d.Id(), protectionOpts)
			if err != nil {
				return diag.Errorf("error updating login protection configuration for IAM user: %s", err)
			}
		} else {
			oldMethod, _ := d.GetChange("login_protection.0.verification_method")
			_, err = security.UpdateLoginProtectionConfiguration(client, d.Id(),
				security.LoginProtectionUpdateOpts{
					Enabled:            pointerto.Bool(false),
					VerificationMethod: oldMethod.(string),
				})
			if err != nil {
				return diag.Errorf("error updating login protection configuration for IAM user: %s", err)
			}
		}
	}

	_, err = users.ModifyUser(client, d.Id(), updateOpts)
	if err != nil {
		return diag.Errorf("error updating IAM user: %s", err)
	}

	if hasChange {
		if d.Get("send_welcome_email").(bool) {
			if err := oldusers.SendWelcomeEmail(client, d.Id()).ExtractErr(); err != nil {
				return fmterr.Errorf("error sending a welcome email: %w", err)
			}
		}
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV3)
	return resourceIdentityUserV3Read(clientCtx, d, meta)
}

func resourceIdentityUserV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV3, func() (*golangsdk.ServiceClient, error) {
		return config.IdentityV3Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(clientCreationFail, err)
	}
	err = checkAndModifyXuser(ctx, d, meta)
	if err != nil {
		diag.FromErr(err)
	}

	err = oldusers.Delete(client, d.Id()).ExtractErr()
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud user: %w", err)
	}

	return nil
}

func normalizePhoneNumber(raw string) string {
	phone := raw

	rawList := strings.Split(raw, "-")
	if len(rawList) > 1 {
		phone = rawList[1]
	}

	return phone
}

func checkAndModifyXuser(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	if d.Get("xuser_type").(string) != "" || d.Get("xuser_id").(string) != "" {
		config := meta.(*cfg.Config)
		client, err := common.ClientFromCtx(ctx, keyClientV30, func() (*golangsdk.ServiceClient, error) {
			return config.IdentityV30Client()
		})
		if err != nil {
			return fmt.Errorf(clientV30CreationFail, err)
		}
		_, err = users.ModifyUserAdmin(client, users.UpdateAdminOpts{
			XuserType: "",
			XuserId:   "",
			Id:        d.Id(),
		})
		if err != nil {
			return fmt.Errorf("error removing xuser_type and xuser_id: %s", err)
		}
		return nil
	}
	return nil
}
