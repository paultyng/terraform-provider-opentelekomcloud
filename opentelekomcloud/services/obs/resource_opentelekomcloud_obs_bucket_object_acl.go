package obs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/obs"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceOBSBucketObjectAcl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOBSBucketObjectAclCreate,
		ReadContext:   resourceOBSBucketObjectAclRead,
		UpdateContext: resourceOBSBucketObjectAclCreate,
		DeleteContext: resourceOBSBucketObjectAclDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOBSBucketObjectAclImportState,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public_permission": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     objectPublicPermissionSchema(),
			},
			"account_permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     objectAccountPermissionSchema(),
			},
			"owner_permission": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     objectOwnerPermissionSchema(),
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func objectPublicPermissionSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"access_to_object": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"READ",
					}, false),
				},
			},
			"access_to_acl": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"READ_ACP", "WRITE_ACP",
					}, false),
				},
			},
		},
	}
	return &sc
}

func objectAccountPermissionSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_to_object": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"READ",
					}, false),
				},
			},
			"access_to_acl": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"READ_ACP", "WRITE_ACP",
					}, false),
				},
			},
		},
	}
	return &sc
}

func objectOwnerPermissionSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"access_to_object": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"access_to_acl": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
	return &sc
}

func buildObjectOwnerPermissionGrants(obsClient *obs.ObsClient, d *schema.ResourceData,
	domainID string) ([]obs.Grant, error) {
	params := &obs.GetObjectAclInput{
		Bucket: d.Get("bucket").(string),
		Key:    d.Get("key").(string),
	}
	output, err := obsClient.GetObjectAcl(params)
	if err != nil {
		return nil, err
	}
	var ownerGrants []obs.Grant
	for _, grant := range output.Grants {
		grantee := grant.Grantee
		if grantee.Type == obs.GranteeUser && grantee.ID == domainID {
			ownerGrants = append(ownerGrants, grant)
		}
	}
	if len(ownerGrants) > 0 {
		return ownerGrants, nil
	}

	accesses := []string{"READ", "READ_ACP", "WRITE_ACP"}
	return buildUserTypeGrants(accesses, domainID), nil
}

func buildObjectAccessesFromRawMap(rawMap map[string]interface{}) []string {
	var accesses []string
	if accessArray, ok := rawMap["access_to_object"].([]interface{}); ok {
		accesses = append(accesses, common.ExpandToStringList(accessArray)...)
	}
	if accessArray, ok := rawMap["access_to_acl"].([]interface{}); ok {
		accesses = append(accesses, common.ExpandToStringList(accessArray)...)
	}
	return accesses
}

func buildObsBucketObjectAclGrants(obsClient *obs.ObsClient, d *schema.ResourceData,
	domainID string) ([]obs.Grant, error) {
	var grants []obs.Grant
	ownerPermissions, err := buildObjectOwnerPermissionGrants(obsClient, d, domainID)
	if err != nil {
		return nil, err
	}
	grants = append(grants, ownerPermissions...)

	permissions := d.Get("account_permission").(*schema.Set)
	for _, raw := range permissions.List() {
		if rawMap, rawOk := raw.(map[string]interface{}); rawOk {
			accountID := rawMap["account_id"].(string)
			if accountID == domainID {
				return nil, fmt.Errorf("the account id cannot be the object owner: %s", accountID)
			}
			accesses := buildObjectAccessesFromRawMap(rawMap)
			log.Printf("[DEBUG] The account permission accesses: %v.", accesses)
			grants = append(grants, buildUserTypeGrants(accesses, accountID)...)
		}
	}

	if rawArray, ok := d.Get("public_permission").([]interface{}); ok && len(rawArray) > 0 {
		if rawMap, rawOk := rawArray[0].(map[string]interface{}); rawOk {
			accesses := buildObjectAccessesFromRawMap(rawMap)
			log.Printf("[DEBUG] The public permission accesses: %v.", accesses)
			grants = append(grants, buildGroupTypeGrants(accesses, obs.GroupAllUsers)...)
		}
	}
	return grants, nil
}

func resourceOBSBucketObjectAclCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	grantParam, err := buildObsBucketObjectAclGrants(client, d, getDomainID(config))
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	params := &obs.SetObjectAclInput{
		Bucket: bucket,
		Key:    key,
	}
	params.Owner.ID = getDomainID(config)
	params.Grants = grantParam
	_, err = client.SetObjectAcl(params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating OBS bucket %s object acl: %s", bucket, err))
	}
	d.SetId(key)
	return resourceOBSBucketObjectAclRead(ctx, d, meta)
}

func flattenObjectAccessesFromGrant(grant obs.Grant) (objectAccesses []string, aclAccesses []string) {
	switch grant.Permission {
	case obs.PermissionRead:
		objectAccesses = []string{"READ"}
	case obs.PermissionReadAcp:
		aclAccesses = []string{"READ_ACP"}
	case obs.PermissionWriteAcp:
		aclAccesses = []string{"WRITE_ACP"}
	case obs.PermissionFullControl:
		objectAccesses = []string{"READ"}
		aclAccesses = []string{"READ_ACP", "WRITE_ACP"}
	default:
		log.Printf("[WARN] The grant permission: %s not support.", grant.Permission)
	}
	return
}

func flattenObjectPermission(grants []obs.Grant) []map[string]interface{} {
	if len(grants) == 0 {
		return nil
	}
	var accessToObject []string
	var accessToAcl []string
	for _, grant := range grants {
		objectAccesses, aclAccesses := flattenObjectAccessesFromGrant(grant)
		accessToObject = append(accessToObject, objectAccesses...)
		accessToAcl = append(accessToAcl, aclAccesses...)
	}
	if len(accessToObject) == 0 && len(accessToAcl) == 0 {
		return nil
	}
	permissionMap := map[string]interface{}{
		"access_to_object": accessToObject,
		"access_to_acl":    accessToAcl,
	}
	return []map[string]interface{}{permissionMap}
}

func flattenObjectAccountPermission(grants []obs.Grant) []map[string]interface{} {
	if len(grants) == 0 {
		return nil
	}
	accountIDSet := make(map[string]bool)
	accessToObjectMap := make(map[string][]string)
	accessToAclMap := make(map[string][]string)
	for _, grant := range grants {
		accountID := grant.Grantee.ID
		objectAccesses, aclAccesses := flattenObjectAccessesFromGrant(grant)
		accessToObjectMap[accountID] = append(accessToObjectMap[accountID], objectAccesses...)
		accessToAclMap[accountID] = append(accessToAclMap[accountID], aclAccesses...)
		accountIDSet[accountID] = true
	}

	var m []map[string]interface{}
	for id := range accountIDSet {
		m = append(m, map[string]interface{}{
			"access_to_object": accessToObjectMap[id],
			"access_to_acl":    accessToAclMap[id],
			"account_id":       id,
		})
	}
	return m
}

func resourceOBSBucketObjectAclRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	params := &obs.GetObjectAclInput{
		Bucket: d.Get("bucket").(string),
		Key:    d.Id(),
	}
	output, err := client.GetObjectAcl(params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving OBS bucket: %s object acl: %s", d.Id(), err))
	}

	permissionTypeMap := flattenGrantsByPermissionType(output.Grants, config)
	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("public_permission", flattenObjectPermission(permissionTypeMap[GrantPublic])),
		d.Set("account_permission", flattenObjectAccountPermission(permissionTypeMap[GrantAccount])),
		d.Set("owner_permission", flattenObjectPermission(permissionTypeMap[GrantOwner])),
	)

	if err = mErr.ErrorOrNil(); err != nil {
		return diag.Errorf("error setting OBS bucket object acl fields: %s", err)
	}
	return nil
}

func resourceOBSBucketObjectAclDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	ownerPermissions, err := buildObjectOwnerPermissionGrants(client, d, getDomainID(config))
	if err != nil {
		return diag.FromErr(err)
	}

	bucket := d.Get("bucket").(string)
	params := &obs.SetObjectAclInput{
		Bucket: bucket,
		Key:    d.Id(),
	}
	params.Owner.ID = getDomainID(config)
	params.Grants = ownerPermissions
	_, err = client.SetObjectAcl(params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting OBS bucket %s object acl: %s", d.Id(), err))
	}
	return nil
}

func resourceOBSBucketObjectAclImportState(_ context.Context, d *schema.ResourceData,
	_ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		err := fmt.Errorf("invalid format specified for import id, must be <bucket>/<key>")
		return nil, err
	}

	bucket := parts[0]
	key := parts[1]
	d.SetId(key)
	mErr := multierror.Append(nil,
		d.Set("bucket", bucket),
		d.Set("key", key),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return nil, fmt.Errorf("failed to set value to state when import obs bucket object acl, %s", err)
	}
	return []*schema.ResourceData{d}, nil
}
