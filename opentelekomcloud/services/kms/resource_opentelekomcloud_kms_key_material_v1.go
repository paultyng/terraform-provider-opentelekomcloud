package kms

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/kms/v1/keys"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceKmsKeyMaterialV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceKmsKeyMaterialCreate,
		ReadContext:   ResourceKmsKeyMaterialRead,
		DeleteContext: ResourceKmsKeyMaterialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"import_token": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encrypted_key_material": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"expiration_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"key_state": {
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

func ResourceKmsKeyMaterialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	importMaterialOpts := keys.ImportCMKOpts{
		KeyId:                d.Get("key_id").(string),
		ImportToken:          d.Get("import_token").(string),
		EncryptedKeyMaterial: d.Get("encrypted_key_material").(string),
		ExpirationTime:       d.Get("expiration_time").(string),
	}

	err = keys.ImportCMKMaterial(client, importMaterialOpts)
	if err != nil {
		return diag.Errorf("error importing KMS key material: %s", err)
	}

	d.SetId(d.Get("key_id").(string))

	return ResourceKmsKeyMaterialRead(ctx, d, meta)
}

func ResourceKmsKeyMaterialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	v, err := keys.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving the KMS key")
	}
	if v.KeyState == PendingDeletionState || v.KeyState == WaitingImportState {
		return common.CheckDeletedDiag(d, err,
			"The KMS key is pending deletion or the key material is pending import")
	}

	expirationTime := flatternExpirationTime(v.ExpirationTime)

	d.SetId(v.KeyID)
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("key_id", v.KeyID),
		d.Set("key_state", v.KeyState),
		d.Set("expiration_time", expirationTime),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func flatternExpirationTime(expTimeStr string) string {
	if expTimeStr == "" {
		return ""
	}
	expTime, _ := strconv.ParseInt(expTimeStr, 10, 64)
	return strconv.FormatInt(expTime/1000, 10)
}

func ResourceKmsKeyMaterialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.KmsKeyV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClient, err)
	}

	v, err := keys.Get(client, d.Id())
	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving the KMS key")
	}
	if v.KeyState == PendingDeletionState || v.KeyState == WaitingImportState {
		return common.CheckDeletedDiag(d, err, "The KMS key is pending deletion or the key material is deleted")
	}

	deleteMaterialOpts := keys.DeleteCMKImportOpts{
		KeyId: d.Id(),
	}

	// The key material of the asymmetric key does not support deletion.
	// Deleting the key material of an asymmetric key will return {"error":{"error_msg":"xx","error_code":"KMS.2702"}}
	// The key material of the symmetric key support deletion.
	err = keys.DeleteCMKImport(client, deleteMaterialOpts)
	if _, ok := err.(golangsdk.ErrDefault400); ok {
		errCode, errMessage := parseDeleteResponseError(err)
		if errCode == "KMS.2702" {
			log.Printf("[WARN] failed to delete key material, errCode : %s, errMsg: %s", errCode, errMessage)
			errorMessage := "The asymmetric key material can't be deleted. The project is only removed from the state," +
				" but it remains in the cloud."
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  errorMessage,
				},
			}
		}
	}
	if err != nil {
		diag.Errorf("error deleting key material (%s): %s", d.Id(), err)
	}

	return nil
}

func parseDeleteResponseError(err error) (errorCode, errorMsg string) {
	var response interface{}
	if jsonErr := json.Unmarshal(err.(golangsdk.ErrDefault400).Body, &response); jsonErr == nil {
		errorCode, parseErr := jmespath.Search("error.error_code", response)
		if parseErr != nil {
			log.Printf("[WARN] failed to parse error_code from response body: %s", parseErr)
		}
		errMsg, parseErr := jmespath.Search("error.error_msg", response)
		if parseErr != nil {
			log.Printf("[WARN] failed to parse error_msg from response body: %s", parseErr)
		}
		return errorCode.(string), errMsg.(string)
	}
	log.Printf("[WARN] failed to parse KMS error message from response body: %s", err)
	return "", ""
}
