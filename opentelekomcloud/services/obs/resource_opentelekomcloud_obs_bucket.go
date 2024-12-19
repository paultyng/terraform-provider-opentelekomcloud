package obs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/obs"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/services/s3"
)

func ResourceObsBucket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObsBucketCreate,
		ReadContext:   resourceObsBucketRead,
		UpdateContext: resourceObsBucketUpdate,
		DeleteContext: resourceObsBucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateVersionObjLock,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateName,
			},
			"storage_class": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "STANDARD",
				ValidateFunc: validation.StringInSlice([]string{
					"STANDARD", "WARM", "COLD",
				}, true),
			},
			"acl": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "private",
				ValidateFunc: validation.StringInSlice([]string{
					"private", "public-read", "public-read-write", "log-delivery-write",
				}, true),
			},
			"versioning": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"parallel_fs": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"logging": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_prefix": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "logs/",
						},
					},
				},
			},
			"worm_policy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"days": {
							Type:     schema.TypeInt,
							Optional: true,
							ConflictsWith: []string{
								"worm_policy.0.years",
							},
							AtLeastOneOf: []string{
								"worm_policy.0.years", "worm_policy.0.days",
							},
						},
						"years": {
							Type:     schema.TypeInt,
							Optional: true,
							ConflictsWith: []string{
								"worm_policy.0.days",
							},
							AtLeastOneOf: []string{
								"worm_policy.0.years", "worm_policy.0.days",
							},
						},
					},
				},
			},
			"lifecycle_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"expiration": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"transition": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"storage_class": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"WARM", "COLD",
										}, true),
									},
								},
							},
						},
						"noncurrent_version_expiration": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"storage_class": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"WARM", "COLD",
										}, true),
									},
								},
							},
						},
					},
				},
			},
			"website": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_document": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"error_document": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"redirect_all_requests_to": {
							Type: schema.TypeString,
							ConflictsWith: []string{
								"website.0.index_document",
								"website.0.error_document",
								"website.0.routing_rules",
							},
							Optional: true,
						},

						"routing_rules": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: common.ValidateJsonString,
							StateFunc: func(v interface{}) string {
								jsonString, _ := common.NormalizeJsonString(v)
								return jsonString
							},
						},
					},
				},
			},
			"cors_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_origins": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  100,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_side_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"algorithm": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"kms"}, false),
							),
						},
					},
				},
			},
			"event_notifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"topic": {
							Type:     schema.TypeString,
							Required: true,
						},
						"id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"events": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"filter_rule": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice(
											[]string{"prefix", "suffix"}, false,
										),
									},
									"value": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
					},
				},
			},
			"user_domain_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"bucket_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceObsBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	bucket := d.Get("bucket").(string)
	acl := d.Get("acl").(string)

	var class string
	if config.GetRegion(d) != "eu-ch2" {
		class = d.Get("storage_class").(string)
	}

	opts := &obs.CreateBucketInput{
		Bucket:            bucket,
		ACL:               obs.AclType(acl),
		IsFSFileInterface: d.Get("parallel_fs").(bool),
		StorageClass:      obs.StorageClassType(class),
	}
	opts.Location = config.GetRegion(d)
	log.Printf("[DEBUG] OBS bucket create opts: %#v", opts)

	if d.Get("worm_policy.#") != 0 {
		opts.ObjectLockEnabled = true
	}

	_, err = client.CreateBucket(opts)
	if err != nil {
		return diag.FromErr(GetObsError("error creating bucket", bucket, err))
	}

	// Assign the bucket name as the resource ID
	d.SetId(bucket)
	return resourceObsBucketUpdate(ctx, d, meta)
}

func resourceObsBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	log.Printf("[DEBUG] Update OBS bucket %s", d.Id())
	if d.HasChange("acl") && !d.IsNewResource() {
		if err := resourceObsBucketAclUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("storage_class") && !d.IsNewResource() && config.GetRegion(d) != "eu-ch2" {
		if err := resourceObsBucketClassUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		if err := resourceObsBucketTagsUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("versioning") {
		versioning := d.Get("versioning").(bool)
		if versioning || !d.IsNewResource() {
			if err := resourceObsBucketVersioningUpdate(client, d); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("logging") {
		if err := resourceObsBucketLoggingUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("lifecycle_rule") {
		if err := resourceObsBucketLifecycleUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("website") {
		if err := resourceObsBucketWebsiteUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("cors_rule") {
		if err := resourceObsBucketCorsUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("server_side_encryption") {
		if err := resourceObsBucketEncryptionUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("event_notifications") {
		if err := resourceObsBucketNotificationUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("worm_policy") {
		if err := resourceObsBucketObjLockUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("user_domain_names") {
		if err := resourceObsBucketUserDomainNamesUpdate(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceObsBucketRead(ctx, d, meta)
}

func resourceObsBucketRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	region := config.GetRegion(d)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	log.Printf("[DEBUG] Read OBS bucket: %s", d.Id())
	_, err = client.HeadBucket(d.Id())
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == 404 {
			log.Printf("[WARN] OBS bucket(%s) not found", d.Id())
			d.SetId("")
			return nil
		} else {
			return fmterr.Errorf("error reading OBS bucket %s: %s", d.Id(), err)
		}
	}

	mErr := &multierror.Error{}

	// for import case
	if _, ok := d.GetOk("bucket"); !ok {
		mErr = multierror.Append(mErr, d.Set("bucket", d.Id()))
	}

	mErr = multierror.Append(mErr,
		d.Set("region", region),
		d.Set("bucket_domain_name", s3.BucketDomainName(d.Get("bucket").(string), region)),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting OBS bucket fields: %s", err)
	}

	// Read storage class
	if region != "eu-ch2" {
		if err := setObsBucketStorageClass(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	// Read the versioning
	if err := setObsBucketVersioning(client, d); err != nil {
		return diag.FromErr(err)
	}
	// Read the logging configuration
	if err := setObsBucketLogging(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read the Lifecycle configuration
	if err := setObsBucketLifecycleConfiguration(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read the website configuration
	if err := setObsBucketWebsiteConfiguration(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read the CORS rules
	if err := setObsBucketCorsRules(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read the tags
	if err := setObsBucketTags(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read SSE settings
	if err := setObsBucketEncryption(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read notifications settings
	if err := setObsBucketNotifications(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read parallel_fs
	if err := setObsBucketMetadata(client, d); err != nil {
		return diag.FromErr(err)
	}

	// Read object lock
	if d.Get("worm_policy.#") != 0 {
		if err := setObsBucketObjLock(client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := setObsBucketUserDomainNames(client, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceObsBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	bucket := d.Id()
	log.Printf("[DEBUG] deleting OBS Bucket: %s", bucket)
	_, err = client.DeleteBucket(bucket)
	if err != nil {
		obsError, ok := err.(obs.ObsError)
		if ok && obsError.Code == "BucketNotEmpty" {
			log.Printf("[WARN] OBS bucket: %s is not empty", bucket)
			if d.Get("force_destroy").(bool) {
				err = deleteAllBucketObjects(client, bucket)
				if err == nil {
					log.Printf("[WARN] all objects of %s have been deleted, and try again", bucket)
					return resourceObsBucketDelete(ctx, d, meta)
				}
			}
			return diag.FromErr(err)
		}
		return fmterr.Errorf("error deleting OBS bucket: %s %s", bucket, err)
	}
	return nil
}

func resourceObsBucketTagsUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	tagMap := d.Get("tags").(map[string]interface{})
	var tagList []obs.Tag
	for k, v := range tagMap {
		tag := obs.Tag{
			Key:   k,
			Value: v.(string),
		}
		tagList = append(tagList, tag)
	}

	req := &obs.SetBucketTaggingInput{}
	req.Bucket = bucket
	req.Tags = tagList
	log.Printf("[DEBUG] set tags of OBS bucket %s: %#v", bucket, req)

	_, err := client.SetBucketTagging(req)
	if err != nil {
		return GetObsError("error updating tags of OBS bucket", bucket, err)
	}
	return nil
}

func resourceObsBucketAclUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	acl := d.Get("acl").(string)

	i := &obs.SetBucketAclInput{
		Bucket: bucket,
		ACL:    obs.AclType(acl),
	}
	log.Printf("[DEBUG] set ACL of OBS bucket %s: %#v", bucket, i)

	_, err := client.SetBucketAcl(i)
	if err != nil {
		return GetObsError("error updating acl of OBS bucket", bucket, err)
	}

	// acl policy can not be retrieved by obsClient.GetBucketAcl method
	err = d.Set("acl", acl)
	return err
}

func resourceObsBucketClassUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	class := d.Get("storage_class").(string)

	input := &obs.SetBucketStoragePolicyInput{}
	input.Bucket = bucket
	input.StorageClass = obs.StorageClassType(class)
	log.Printf("[DEBUG] set storage class of OBS bucket %s: %#v", bucket, input)

	_, err := client.SetBucketStoragePolicy(input)
	if err != nil {
		return GetObsError("error updating storage class of OBS bucket", bucket, err)
	}

	return nil
}

func resourceObsBucketObjLockUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	days := d.Get("worm_policy.0.days").(int)
	years := d.Get("worm_policy.0.years").(int)

	input := obs.SetWORMPolicyInput{
		Bucket: bucket,
		BucketWormPolicy: obs.BucketWormPolicy{
			ObjectLockEnabled: "Enabled",
			Mode:              "COMPLIANCE",
		},
	}

	if days != 0 {
		input.Days = strconv.Itoa(days)
	} else {
		input.Years = strconv.Itoa(years)
	}

	log.Printf("[DEBUG] set object lock of OBS bucket %s: %#v", bucket, input)

	_, err := client.SetWORMPolicy(&input)

	if err != nil {
		return GetObsError("error updating object lock of OBS bucket", bucket, err)
	}

	return nil
}

func resourceObsBucketVersioningUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	versioning := d.Get("versioning").(bool)

	input := &obs.SetBucketVersioningInput{}
	input.Bucket = bucket
	if versioning {
		input.Status = obs.VersioningStatusEnabled
	} else {
		input.Status = obs.VersioningStatusSuspended
	}
	log.Printf("[DEBUG] set versioning of OBS bucket %s: %#v", bucket, input)

	_, err := client.SetBucketVersioning(input)
	if err != nil {
		return GetObsError("error setting versioning status of OBS bucket", bucket, err)
	}

	return nil
}

func resourceObsBucketLoggingUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	rawLogging := d.Get("logging").(*schema.Set).List()
	loggingStatus := &obs.SetBucketLoggingConfigurationInput{}
	loggingStatus.Bucket = bucket

	if len(rawLogging) > 0 {
		c := rawLogging[0].(map[string]interface{})
		if val := c["target_bucket"].(string); val != "" {
			loggingStatus.TargetBucket = val
		}

		if val := c["target_prefix"].(string); val != "" {
			loggingStatus.TargetPrefix = val
		}
	}
	log.Printf("[DEBUG] set logging of OBS bucket %s: %#v", bucket, loggingStatus)

	_, err := client.SetBucketLoggingConfiguration(loggingStatus)
	if err != nil {
		return GetObsError("error setting logging configuration of OBS bucket", bucket, err)
	}

	return nil
}

func mapToRule(src map[string]interface{}) (rule obs.LifecycleRule) {
	// rule ID
	rule.ID = src["name"].(string)

	// Enabled
	if val, ok := src["enabled"].(bool); ok && val {
		rule.Status = obs.RuleStatusEnabled
	} else {
		rule.Status = obs.RuleStatusDisabled
	}

	// Prefix
	rule.Prefix = src["prefix"].(string)

	// Expiration
	expiration := src["expiration"].(*schema.Set).List()
	if len(expiration) > 0 {
		raw := expiration[0].(map[string]interface{})
		exp := &rule.Expiration

		if val, ok := raw["days"].(int); ok && val > 0 {
			exp.Days = val
		}
	}

	// Transition
	transitions := src["transition"].([]interface{})
	list := make([]obs.Transition, len(transitions))
	for j, tran := range transitions {
		raw := tran.(map[string]interface{})

		if val, ok := raw["days"].(int); ok && val > 0 {
			list[j].Days = val
		}
		if val, ok := raw["storage_class"].(string); ok {
			list[j].StorageClass = obs.StorageClassType(val)
		}
	}
	rule.Transitions = list

	// NoncurrentVersionExpiration
	ncExpiration := src["noncurrent_version_expiration"].(*schema.Set).List()
	if len(ncExpiration) > 0 {
		raw := ncExpiration[0].(map[string]interface{})
		ncExp := &rule.NoncurrentVersionExpiration

		if val, ok := raw["days"].(int); ok && val > 0 {
			ncExp.NoncurrentDays = val
		}
	}

	// NoncurrentVersionTransition
	ncTransitions := src["noncurrent_version_transition"].([]interface{})
	ncList := make([]obs.NoncurrentVersionTransition, len(ncTransitions))
	for j, ncTran := range ncTransitions {
		raw := ncTran.(map[string]interface{})

		if val, ok := raw["days"].(int); ok && val > 0 {
			ncList[j].NoncurrentDays = val
		}
		if val, ok := raw["storage_class"].(string); ok {
			ncList[j].StorageClass = obs.StorageClassType(val)
		}
	}
	rule.NoncurrentVersionTransitions = ncList
	return rule
}

func resourceObsBucketLifecycleUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	lifecycleRules := d.Get("lifecycle_rule").([]interface{})

	if len(lifecycleRules) == 0 {
		log.Printf("[DEBUG] remove all lifecycle rules of bucket %s", bucket)
		_, err := client.DeleteBucketLifecycleConfiguration(bucket)
		if err != nil {
			return GetObsError("error deleting lifecycle rules of OBS bucket", bucket, err)
		}
		return nil
	}

	rules := make([]obs.LifecycleRule, len(lifecycleRules))
	for i, lifecycleRule := range lifecycleRules {
		ruleMap := lifecycleRule.(map[string]interface{})
		rules[i] = mapToRule(ruleMap)
	}

	opts := &obs.SetBucketLifecycleConfigurationInput{}
	opts.Bucket = bucket
	opts.LifecycleRules = rules
	log.Printf("[DEBUG] set lifecycle configurations of OBS bucket %s: %#v", bucket, opts)

	_, err := client.SetBucketLifecycleConfiguration(opts)
	if err != nil {
		return GetObsError("error setting lifecycle rules of OBS bucket", bucket, err)
	}

	return nil
}

func resourceObsBucketWebsiteUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	websiteList := d.Get("website").([]interface{})

	switch len(websiteList) {
	case 0:
		return resourceObsBucketWebsiteDelete(client, d)
	case 1:
		var website map[string]interface{}
		if websiteList[0] != nil {
			website = websiteList[0].(map[string]interface{})
		} else {
			website = make(map[string]interface{})
		}
		return resourceObsBucketWebsitePut(client, d, website)
	default:
		return fmt.Errorf("cannot specify more than one website")
	}
}

// asStringSlice returns `value` and `ok`
func asStringSlice(value interface{}) ([]string, bool) {
	sliceVal, ok := value.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]string, len(sliceVal))
	for i, val := range sliceVal {
		result[i], ok = val.(string)
		if !ok {
			return nil, false
		}
	}
	return result, true
}

func resourceObsBucketCorsUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) == 0 {
		// Delete CORS
		log.Printf("[DEBUG] delete CORS rules of OBS bucket: %s", bucket)
		_, err := client.DeleteBucketCors(bucket)
		if err != nil {
			return GetObsError("error deleting CORS rules of OBS bucket", bucket, err)
		}
		return nil
	}

	// set CORS
	rules := make([]obs.CorsRule, len(rawCors))
	for i, cors := range rawCors {
		corsMap := cors.(map[string]interface{})
		rule := obs.CorsRule{}
		for k, v := range corsMap {
			if k == "max_age_seconds" {
				rule.MaxAgeSeconds = v.(int)
				continue
			}
			value, ok := asStringSlice(v)
			if !ok {
				continue
			}
			switch k {
			case "allowed_headers":
				rule.AllowedHeader = value
			case "allowed_methods":
				rule.AllowedMethod = value
			case "allowed_origins":
				rule.AllowedOrigin = value
			case "expose_headers":
				rule.ExposeHeader = value
			}
		}
		log.Printf("[DEBUG] set CORS of OBS bucket %s: %#v", bucket, rule)
		rules[i] = rule
	}

	corsInput := &obs.SetBucketCorsInput{}
	corsInput.Bucket = bucket
	corsInput.CorsRules = rules
	log.Printf("[DEBUG] OBS bucket: %s, put CORS: %#v", bucket, corsInput)

	_, err := client.SetBucketCors(corsInput)
	if err != nil {
		return GetObsError("error setting CORS rules of OBS bucket", bucket, err)
	}
	return nil
}

func resourceObsBucketWebsitePut(client *obs.ObsClient, d *schema.ResourceData, website map[string]interface{}) error {
	bucket := d.Get("bucket").(string)

	var indexDocument, errorDocument, redirectAllRequestsTo, routingRules string
	if v, ok := website["index_document"]; ok {
		indexDocument = v.(string)
	}
	if v, ok := website["error_document"]; ok {
		errorDocument = v.(string)
	}
	if v, ok := website["redirect_all_requests_to"]; ok {
		redirectAllRequestsTo = v.(string)
	}
	if v, ok := website["routing_rules"]; ok {
		routingRules = v.(string)
	}

	if indexDocument == "" && redirectAllRequestsTo == "" {
		return fmt.Errorf("must specify either index_document or redirect_all_requests_to")
	}

	websiteConfiguration := &obs.SetBucketWebsiteConfigurationInput{}
	websiteConfiguration.Bucket = bucket

	if indexDocument != "" {
		websiteConfiguration.IndexDocument = obs.IndexDocument{
			Suffix: indexDocument,
		}
	}

	if errorDocument != "" {
		websiteConfiguration.ErrorDocument = obs.ErrorDocument{
			Key: errorDocument,
		}
	}

	if redirectAllRequestsTo != "" {
		redirect, err := url.Parse(redirectAllRequestsTo)
		if err == nil && redirect.Scheme != "" {
			var redirectHostBuf bytes.Buffer
			redirectHostBuf.WriteString(redirect.Host)
			if redirect.Path != "" {
				redirectHostBuf.WriteString(redirect.Path)
			}
			websiteConfiguration.RedirectAllRequestsTo = obs.RedirectAllRequestsTo{
				HostName: redirectHostBuf.String(),
				Protocol: obs.ProtocolType(redirect.Scheme),
			}
		} else {
			websiteConfiguration.RedirectAllRequestsTo = obs.RedirectAllRequestsTo{
				HostName: redirectAllRequestsTo,
			}
		}
	}

	if routingRules != "" {
		var unmarshalledRules []obs.RoutingRule
		if err := json.Unmarshal([]byte(routingRules), &unmarshalledRules); err != nil {
			return err
		}
		websiteConfiguration.RoutingRules = unmarshalledRules
	}

	log.Printf("[DEBUG] set website configuration of OBS bucket %s: %#v", bucket, websiteConfiguration)
	_, err := client.SetBucketWebsiteConfiguration(websiteConfiguration)
	if err != nil {
		return GetObsError("error updating website configuration of OBS bucket", bucket, err)
	}

	return nil
}

func resourceObsBucketWebsiteDelete(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)

	log.Printf("[DEBUG] delete website configuration of OBS bucket %s", bucket)
	_, err := client.DeleteBucketWebsiteConfiguration(bucket)
	if err != nil {
		return GetObsError("error deleting website configuration of OBS bucket", bucket, err)
	}

	return nil
}

func setObsBucketStorageClass(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := client.GetBucketStoragePolicy(bucket)
	if err != nil {
		return GetObsError("error getting storage class of OBS bucket", bucket, err)
	}

	class := output.StorageClass
	err = d.Set("storage_class", normalizeStorageClass(class))
	return err
}

func setObsBucketVersioning(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := client.GetBucketVersioning(bucket)
	if err != nil {
		return GetObsError("error getting versioning status of OBS bucket", bucket, err)
	}

	enabled := output.Status == obs.VersioningStatusEnabled
	err = d.Set("versioning", enabled)

	return err
}

func setObsBucketLogging(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := client.GetBucketLoggingConfiguration(bucket)
	if err != nil {
		return GetObsError("error getting logging configuration of OBS bucket", bucket, err)
	}

	var lcList []map[string]interface{}
	logging := make(map[string]interface{})

	if output.TargetBucket != "" {
		logging["target_bucket"] = output.TargetBucket
		if output.TargetPrefix != "" {
			logging["target_prefix"] = output.TargetPrefix
		}
		lcList = append(lcList, logging)
	}
	log.Printf("[DEBUG] saving logging configuration of OBS bucket: %s: %#v", bucket, lcList)

	if err := d.Set("logging", lcList); err != nil {
		return fmt.Errorf("error saving logging configuration of OBS bucket %s: %s", bucket, err)
	}

	return nil
}

func ruleToMap(src obs.LifecycleRule) map[string]interface{} {
	rule := make(map[string]interface{})
	rule["name"] = src.ID

	// Enabled
	rule["enabled"] = src.Status == obs.RuleStatusEnabled

	if src.Prefix != "" {
		rule["prefix"] = src.Prefix
	}

	// expiration
	if days := src.Expiration.Days; days > 0 {
		expiration := make(map[string]interface{})
		expiration["days"] = days
		rule["expiration"] = schema.NewSet(s3.ExpirationHash, []interface{}{expiration})
	}
	// transition
	if len(src.Transitions) > 0 {
		transitions := make([]interface{}, len(src.Transitions))
		for i, v := range src.Transitions {
			transition := make(map[string]interface{})
			transition["days"] = v.Days
			transition["storage_class"] = normalizeStorageClass(string(v.StorageClass))
			transitions[i] = transition
		}
		rule["transition"] = transitions
	}

	// noncurrent_version_expiration
	if days := src.NoncurrentVersionExpiration.NoncurrentDays; days > 0 {
		expiration := make(map[string]interface{})
		expiration["days"] = days
		rule["noncurrent_version_expiration"] = schema.NewSet(s3.ExpirationHash, []interface{}{expiration})
	}

	// noncurrent_version_transition
	if len(src.NoncurrentVersionTransitions) > 0 {
		transitions := make([]interface{}, len(src.NoncurrentVersionTransitions))
		for i, v := range src.NoncurrentVersionTransitions {
			transition := make(map[string]interface{})
			transition["days"] = v.NoncurrentDays
			transition["storage_class"] = normalizeStorageClass(string(v.StorageClass))
			transitions[i] = transition
		}
		rule["noncurrent_version_transition"] = transitions
	}
	return rule
}

func setObsBucketLifecycleConfiguration(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := client.GetBucketLifecycleConfiguration(bucket)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			if obsError.Code == "NoSuchLifecycleConfiguration" {
				err = d.Set("lifecycle_rule", nil)
				return err
			}
			return fmt.Errorf("error getting lifecycle configuration of OBS bucket %s: %s,\n Reason: %s",
				bucket, obsError.Code, obsError.Message)
		}
		return err
	}

	rawRules := output.LifecycleRules
	log.Printf("[DEBUG] getting lifecycle configuration of OBS bucket: %s, lifecycle: %#v", bucket, rawRules)

	rules := make([]map[string]interface{}, len(rawRules))
	for i, lifecycleRule := range rawRules {
		rules[i] = ruleToMap(lifecycleRule)
	}

	log.Printf("[DEBUG] saving lifecycle configuration of OBS bucket: %s, lifecycle: %#v", bucket, rules)
	if err := d.Set("lifecycle_rule", rules); err != nil {
		return fmt.Errorf("error saving lifecycle configuration of OBS bucket %s: %s", bucket, err)
	}

	return nil
}

func handleWebsite(src *obs.GetBucketWebsiteConfigurationOutput) (map[string]interface{}, error) {
	website := make(map[string]interface{})
	website["index_document"] = src.IndexDocument.Suffix
	website["error_document"] = src.ErrorDocument.Key

	// redirect_all_requests_to
	v := src.RedirectAllRequestsTo
	if string(v.Protocol) == "" {
		website["redirect_all_requests_to"] = v.HostName
	} else {
		var host string
		var path string
		parsedHostName, err := url.Parse(v.HostName)
		if err == nil {
			host = parsedHostName.Host
			path = parsedHostName.Path
		} else {
			host = v.HostName
			path = ""
		}

		website["redirect_all_requests_to"] = (&url.URL{
			Host:   host,
			Path:   path,
			Scheme: string(v.Protocol),
		}).String()
	}

	// routing_rules
	rawRules := src.RoutingRules
	if len(rawRules) > 0 {
		rr, err := normalizeWebsiteRoutingRules(rawRules)
		if err != nil {
			return nil, fmt.Errorf("error while marshaling website routing rules: %s", err)
		}
		website["routing_rules"] = rr
	}
	return website, nil
}

func setObsBucketWebsiteConfiguration(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	rawWebsite, err := client.GetBucketWebsiteConfiguration(bucket)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			if obsError.Code == "NoSuchWebsiteConfiguration" {
				err = d.Set("website", nil)
				return err
			} else {
				return fmt.Errorf("error getting website configuration of OBS bucket %s: %s,\n Reason: %s",
					bucket, obsError.Code, obsError.Message)
			}
		} else {
			return err
		}
	}

	log.Printf("[DEBUG] getting website configuration of OBS bucket: %s, output: %#v", bucket, rawWebsite.BucketWebsiteConfiguration)

	website, err := handleWebsite(rawWebsite)
	if err != nil {
		return err
	}

	websites := []map[string]interface{}{website}
	log.Printf("[DEBUG] saving website configuration of OBS bucket: %s, website: %#v", bucket, websites)
	if err := d.Set("website", websites); err != nil {
		return fmt.Errorf("error saving website configuration of OBS bucket %s: %s", bucket, err)
	}
	return nil
}

func setObsBucketCorsRules(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := client.GetBucketCors(bucket)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			if obsError.Code == "NoSuchCORSConfiguration" {
				err = d.Set("cors_rule", nil)
				return err
			} else {
				return fmt.Errorf("error getting CORS configuration of OBS bucket %s: %s,\n Reason: %s",
					bucket, obsError.Code, obsError.Message)
			}
		} else {
			return err
		}
	}

	corsRules := output.CorsRules
	log.Printf("[DEBUG] getting CORS rules of OBS bucket: %s, CORS: %#v", bucket, corsRules)

	rules := make([]map[string]interface{}, len(corsRules))
	for i, ruleObject := range corsRules {
		rule := make(map[string]interface{})
		rule["allowed_origins"] = ruleObject.AllowedOrigin
		rule["allowed_methods"] = ruleObject.AllowedMethod
		rule["max_age_seconds"] = ruleObject.MaxAgeSeconds
		if ruleObject.AllowedHeader != nil {
			rule["allowed_headers"] = ruleObject.AllowedHeader
		}
		if ruleObject.ExposeHeader != nil {
			rule["expose_headers"] = ruleObject.ExposeHeader
		}
		rules[i] = rule
	}

	log.Printf("[DEBUG] saving CORS rules of OBS bucket: %s, CORS: %#v", bucket, rules)
	if err := d.Set("cors_rule", rules); err != nil {
		return fmt.Errorf("error saving CORS rules of OBS bucket %s: %s", bucket, err)
	}

	return nil
}

func setObsBucketTags(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := client.GetBucketTagging(bucket)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			if obsError.Code == "NoSuchTagSet" {
				err = d.Set("tags", nil)
				return err
			} else {
				return fmt.Errorf("error getting tags of OBS bucket %s: %s,\n Reason: %s",
					bucket, obsError.Code, obsError.Message)
			}
		} else {
			return err
		}
	}

	tagMap := make(map[string]string)
	for _, tag := range output.Tags {
		tagMap[tag.Key] = tag.Value
	}
	if err := d.Set("tags", tagMap); err != nil {
		return fmt.Errorf("error saving tags of OBS bucket %s: %s", bucket, err)
	}
	return nil
}

func deleteAllBucketObjects(client *obs.ObsClient, bucket string) error {
	// Query and delete objects in cycle because ListVersions returns
	// max of 1000 objects.
	for {
		listOpts := &obs.ListVersionsInput{
			Bucket: bucket,
		}
		// list all objects
		resp, err := client.ListVersions(listOpts)
		if err != nil {
			return GetObsError("error listing objects of OBS bucket", bucket, err)
		}
		objLength := len(resp.DeleteMarkers) + len(resp.Versions)
		objects := make([]obs.ObjectToDelete, objLength)
		for i, content := range resp.Versions {
			objects[i].Key = content.Key
			objects[i].VersionId = content.VersionId
		}
		for i, content := range resp.DeleteMarkers {
			objects[len(resp.Versions)+i].Key = content.Key
			objects[len(resp.Versions)+i].VersionId = content.VersionId
		}

		deleteOpts := &obs.DeleteObjectsInput{
			Bucket:  bucket,
			Objects: objects,
		}
		log.Printf("[DEBUG] objects of %s will be deleted: %v", bucket, objects)
		output, err := client.DeleteObjects(deleteOpts)
		if err != nil {
			return GetObsError("error deleting all objects of OBS bucket", bucket, err)
		} else if len(output.Errors) > 0 {
			return fmt.Errorf("error some objects still exist in %s: %#v", bucket, output.Errors)
		}
		if objLength < 1000 {
			break
		}
	}
	return nil
}

func GetObsError(action string, bucket string, err error) error {
	if obsError, ok := err.(obs.ObsError); ok {
		return fmt.Errorf("%s %s: %s,\n Reason: %s", action, bucket, obsError.Code, obsError.Message)
	}
	return err
}

// normalize format of storage class
func normalizeStorageClass(class string) string {
	switch class {
	case "STANDARD_IA":
		return "WARM"
	case "GLACIER":
		return "COLD"
	default:
		return class
	}
}

func normalizeWebsiteRoutingRules(w []obs.RoutingRule) (string, error) {
	// transform []obs.RoutingRule to []WebsiteRoutingRule
	wRules := make([]WebsiteRoutingRule, len(w))
	for i, rawRule := range w {
		rule := WebsiteRoutingRule{
			Condition: Condition{
				KeyPrefixEquals:             rawRule.Condition.KeyPrefixEquals,
				HttpErrorCodeReturnedEquals: rawRule.Condition.HttpErrorCodeReturnedEquals,
			},
			Redirect: Redirect{
				Protocol:             string(rawRule.Redirect.Protocol),
				HostName:             rawRule.Redirect.HostName,
				HttpRedirectCode:     rawRule.Redirect.HttpRedirectCode,
				ReplaceKeyWith:       rawRule.Redirect.ReplaceKeyWith,
				ReplaceKeyPrefixWith: rawRule.Redirect.ReplaceKeyPrefixWith,
			},
		}
		wRules[i] = rule
	}

	// normalize
	withNulls, err := json.Marshal(wRules)
	if err != nil {
		return "", err
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(withNulls, &rules); err != nil {
		return "", err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, s3.RemoveNil(rule))
	}

	withoutNulls, err := json.Marshal(cleanRules)
	if err != nil {
		return "", err
	}

	return string(withoutNulls), nil
}

type Condition struct {
	KeyPrefixEquals             string `json:"KeyPrefixEquals,omitempty"`
	HttpErrorCodeReturnedEquals string `json:"HttpErrorCodeReturnedEquals,omitempty"`
}

type Redirect struct {
	Protocol             string `json:"Protocol,omitempty"`
	HostName             string `json:"HostName,omitempty"`
	ReplaceKeyPrefixWith string `json:"ReplaceKeyPrefixWith,omitempty"`
	ReplaceKeyWith       string `json:"ReplaceKeyWith,omitempty"`
	HttpRedirectCode     string `json:"HttpRedirectCode,omitempty"`
}

type WebsiteRoutingRule struct {
	Condition Condition `json:"Condition,omitempty"`
	Redirect  Redirect  `json:"Redirect"`
}

func resourceObsBucketEncryptionUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	if d.Get("server_side_encryption.#") == 0 && !d.IsNewResource() {
		_, err := client.DeleteBucketEncryption(d.Id())
		if err != nil {
			return fmt.Errorf("failed to disable default encryption of OBS bucket %s", d.Id())
		}
		return nil
	}
	_, err := client.SetBucketEncryption(&obs.SetBucketEncryptionInput{
		Bucket: d.Id(),
		BucketEncryptionConfiguration: obs.BucketEncryptionConfiguration{
			SSEAlgorithm:   d.Get("server_side_encryption.0.algorithm").(string),
			KMSMasterKeyID: d.Get("server_side_encryption.0.kms_key_id").(string),
		},
	})
	if err != nil {
		return fmt.Errorf("error setting bucket encryption: %w", err)
	}
	return nil
}

func setObsBucketEncryption(client *obs.ObsClient, d *schema.ResourceData) error {
	config, err := client.GetBucketEncryption(d.Id())
	if err != nil {
		if oErr, ok := err.(obs.ObsError); ok {
			if oErr.BaseModel.StatusCode == 404 || oErr.Code == "NoSuchEncryptionConfiguration" || oErr.Code == "FsNotSupport" {
				return nil
			}
		}
		return fmt.Errorf("error reading bucket encryption: %w", err)
	}
	value := []map[string]interface{}{{
		"kms_key_id": config.KMSMasterKeyID,
		"algorithm":  config.SSEAlgorithm,
	}}
	return d.Set("server_side_encryption", value)
}

func resourceObsBucketNotificationUpdate(client *obs.ObsClient, d *schema.ResourceData) error {
	notifications := d.Get("event_notifications").([]interface{})

	configs := make([]obs.TopicConfiguration, len(notifications))
	for i, n := range notifications {
		notification := n.(map[string]interface{})
		config := obs.TopicConfiguration{
			Topic:       notification["topic"].(string),
			ID:          notification["id"].(string),
			Events:      toEventSlice(notification["events"]),
			FilterRules: toFilterRules(notification["filter_rule"]),
		}
		configs[i] = config
	}

	opts := &obs.SetBucketNotificationInput{
		Bucket:             d.Get("bucket").(string),
		BucketNotification: obs.BucketNotification{TopicConfigurations: configs},
	}

	if _, err := client.SetBucketNotification(opts); err != nil {
		return fmt.Errorf("error setting notification for bucket: %w", err)
	}
	return nil
}

func setObsBucketNotifications(client *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	notifications, err := client.GetBucketNotification(bucket)
	if err != nil {
		return fmt.Errorf("error reading bucket notification configuration: %w", err)
	}

	configs := make([]interface{}, len(notifications.TopicConfigurations))
	for i, v := range notifications.TopicConfigurations {
		configs[i] = map[string]interface{}{
			"topic":       v.Topic,
			"id":          v.ID,
			"events":      eventsToStrSlice(v.Events),
			"filter_rule": filterRulesToMapSlice(v.FilterRules),
		}
	}
	return d.Set("event_notifications", configs)
}

func setObsBucketMetadata(obsClient *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	input := &obs.GetBucketMetadataInput{
		Bucket: bucket,
	}
	output, err := obsClient.GetBucketMetadata(input)
	if err != nil {
		return fmt.Errorf("error getting metadata of OBS bucket '%s': %w", bucket, err)
	}
	log.Printf("[DEBUG] getting metadata of OBS bucket %s: %#v", bucket, output)

	mErr := multierror.Append(
		d.Set("bucket_version", output.Version),
	)

	if output.FSStatus == "Enabled" {
		mErr = multierror.Append(mErr,
			d.Set("parallel_fs", true),
		)
	} else {
		mErr = multierror.Append(mErr,
			d.Set("parallel_fs", false),
		)
	}
	if mErr.ErrorOrNil() != nil {
		return fmt.Errorf("error saving metadata of OBS bucket %s: %s", bucket, mErr)
	}
	return nil
}

func setObsBucketObjLock(obsClient *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := obsClient.GetWORMPolicy(bucket)

	if err != nil {
		return fmt.Errorf("error getting object lock of OBS bucket '%s': %w", bucket, err)
	}
	log.Printf("[DEBUG] getting object lock info of OBS bucket %s: %#v", bucket, output)

	var objLock []map[string]interface{}
	theLock := make(map[string]interface{}, 1)

	if output.Days != "" && output.Days != "0" {
		days, _ := strconv.Atoi(output.Days)
		theLock["days"] = days
	} else {
		years, _ := strconv.Atoi(output.Years)
		theLock["years"] = years
	}
	objLock = append(objLock, theLock)

	log.Printf("[DEBUG] saving lobject lock info of OBS bucket: %s: %#v", bucket, objLock)

	if err := d.Set("worm_policy", objLock); err != nil {
		return fmt.Errorf("error saving worm policy of OBS bucket %s: %s", bucket, err)
	}

	return nil
}

func filterRulesToMapSlice(src []obs.FilterRule) []interface{} {
	res := make([]interface{}, len(src))
	for i, v := range src {
		res[i] = map[string]interface{}{
			"name":  v.Name,
			"value": v.Value,
		}
	}
	return res
}

func toFilterRules(src interface{}) []obs.FilterRule {
	rules := src.(*schema.Set)
	res := make([]obs.FilterRule, rules.Len())
	for i, v := range rules.List() {
		rule := v.(map[string]interface{})
		res[i] = obs.FilterRule{
			Name:  rule["name"].(string),
			Value: rule["value"].(string),
		}
	}
	return res
}

func eventsToStrSlice(src []obs.EventType) []string {
	res := make([]string, len(src))
	for i, v := range src {
		res[i] = string(v)
	}
	return res
}

func toEventSlice(src interface{}) []obs.EventType {
	events := src.(*schema.Set)
	res := make([]obs.EventType, events.Len())
	for i, v := range events.List() {
		res[i] = obs.EventType(v.(string))
	}
	return res
}

func validateVersionObjLock(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if d.Get("worm_policy.#") != 0 {
		versioning := d.Get("versioning").(bool)
		if !versioning {
			return fmt.Errorf("OBS requires versioning to be enabled in order to use WORM policies")
		}
	}

	return nil
}

func resourceObsBucketUserDomainNamesUpdate(obsClient *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	oldRaws, newRaws := d.GetChange("user_domain_names")
	addRaws := newRaws.(*schema.Set).Difference(oldRaws.(*schema.Set))
	removeRaws := oldRaws.(*schema.Set).Difference(newRaws.(*schema.Set))

	if err := deleteObsBucketUserDomainNames(obsClient, bucket, removeRaws); err != nil {
		return err
	}
	return createObsBucketUserDomainNames(obsClient, bucket, addRaws)
}

func deleteObsBucketUserDomainNames(obsClient *obs.ObsClient, bucket string, domainNameSet *schema.Set) error {
	for _, domainName := range domainNameSet.List() {
		input := &obs.DeleteBucketCustomDomainInput{
			Bucket:       bucket,
			CustomDomain: domainName.(string),
		}
		_, err := obsClient.DeleteBucketCustomDomain(input)
		if err != nil {
			return fmt.Errorf("error deleting user domain name of OBS bucket %s: %s", bucket, err)
		}
	}
	return nil
}

func createObsBucketUserDomainNames(obsClient *obs.ObsClient, bucket string, domainNameSet *schema.Set) error {
	for _, domainName := range domainNameSet.List() {
		input := &obs.SetBucketCustomDomainInput{
			Bucket:       bucket,
			CustomDomain: domainName.(string),
		}
		_, err := obsClient.SetBucketCustomDomain(input)
		if err != nil {
			return fmt.Errorf("error setting user domain name of OBS bucket %s: %s", bucket, err)
		}
	}
	return nil
}

func setObsBucketUserDomainNames(obsClient *obs.ObsClient, d *schema.ResourceData) error {
	bucket := d.Id()
	output, err := obsClient.GetBucketCustomDomain(bucket)
	if err != nil {
		return fmt.Errorf("error getting user domain names of OBS bucket %s %s", bucket, err)
	}
	log.Printf("[DEBUG] getting user domain names of OBS bucket %s: %#v", bucket, output)

	domainNames := make([]string, len(output.Domains))
	for i, v := range output.Domains {
		domainNames[i] = v.DomainName
	}
	return d.Set("user_domain_names", domainNames)
}
