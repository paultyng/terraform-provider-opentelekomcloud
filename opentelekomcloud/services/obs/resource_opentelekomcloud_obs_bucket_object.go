package obs

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/obs"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceObsBucketObject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObsBucketObjectPut,
		ReadContext:   resourceObsBucketObjectRead,
		UpdateContext: resourceObsBucketObjectPut,
		DeleteContext: resourceObsBucketObjectDelete,

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
			"source": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"content"},
			},
			"content": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"source"},
			},
			"storage_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"STANDARD", "WARM", "COLD",
				}, true),
			},
			"acl": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"private", "public-read", "public-read-write",
				}, true),
			},
			"encryption": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type: schema.TypeString,
				// This will conflict with server-side-encryption and multi-part upload
				// if/when it's actually implemented. The Etag then won't match raw-file MD5.
				Optional: true,
				Computed: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceObsBucketObjectPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var resp *obs.PutObjectOutput
	var err error

	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	if source, ok := d.GetOk("source"); ok {
		// check source file whether exist
		if _, err := os.Stat(source.(string)); err != nil {
			if os.IsNotExist(err) {
				return fmterr.Errorf("source file %s does not exist", source)
			}
			return diag.FromErr(err)
		}

		// put source file
		resp, err = putFileToObject(client, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := d.GetOk("content"); ok {
		// put content
		resp, err = putContentToObject(client, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	if err != nil {
		return diag.FromErr(GetObsError("error putting object to OBS bucket", bucket, err))
	}

	log.Printf("[DEBUG] Response of putting %s to OBS Bucket %s: %#v", key, bucket, resp)
	if resp.VersionId != "null" {
		err = d.Set("version_id", resp.VersionId)
	} else {
		err = d.Set("version_id", "")
	}
	if err != nil {
		return fmterr.Errorf("error setting version_id: %s", err)
	}

	d.SetId(key)

	return resourceObsBucketObjectRead(ctx, d, meta)
}

func basicInput(d *schema.ResourceData) obs.PutObjectBasicInput {
	common := obs.PutObjectBasicInput{
		ObjectOperationInput: obs.ObjectOperationInput{
			Bucket: d.Get("bucket").(string),
			Key:    d.Get("key").(string),
		},
		ContentType: d.Get("content_type").(string),
	}
	if v, ok := d.GetOk("acl"); ok {
		common.ACL = obs.AclType(v.(string))
	}
	if v, ok := d.GetOk("storage_class"); ok {
		common.StorageClass = obs.StorageClassType(v.(string))
	}
	if v, ok := d.GetOk("content_type"); ok {
		common.ContentType = v.(string)
	}
	if d.Get("encryption").(bool) {
		common.SseHeader = obs.SseKmsHeader{
			Encryption: obs.DEFAULT_SSE_KMS_ENCRYPTION_OBS,
			Key:        d.Get("kms_key_id").(string),
		}
	}
	return common
}

func putContentToObject(obsClient *obs.ObsClient, d *schema.ResourceData) (*obs.PutObjectOutput, error) {
	content := d.Get("content").(string)

	putInput := &obs.PutObjectInput{
		PutObjectBasicInput: basicInput(d),
	}

	log.Printf("[DEBUG] putting %s to OBS Bucket %s, opts: %#v", putInput.Key, putInput.Bucket, putInput)
	// do not log content
	body := bytes.NewReader([]byte(content))
	putInput.Body = body

	return obsClient.PutObject(putInput)
}

func putFileToObject(obsClient *obs.ObsClient, d *schema.ResourceData) (*obs.PutObjectOutput, error) {
	putInput := &obs.PutFileInput{
		PutObjectBasicInput: basicInput(d),
	}
	putInput.SourceFile = d.Get("source").(string)

	log.Printf("[DEBUG] putting %s to OBS Bucket %s, opts: %#v", putInput.Key, putInput.Bucket, putInput)
	return obsClient.PutFile(putInput)
}

func resourceObsBucketObjectRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	input := &obs.ListObjectsInput{}
	input.Bucket = bucket
	input.Prefix = key

	resp, err := client.ListObjects(input)
	if err != nil {
		return diag.FromErr(GetObsError("error listing objects of OBS bucket", bucket, err))
	}

	var exist bool
	var object obs.Content
	for _, content := range resp.Contents {
		if key == content.Key {
			exist = true
			object = content
			break
		}
	}
	if !exist {
		d.SetId("")
		return fmterr.Errorf("object %s not found in bucket %s", key, bucket)
	}
	log.Printf("[DEBUG] Reading OBS Bucket Object %s: %#v", key, object)

	if class := string(object.StorageClass); class == "" {
		err = d.Set("storage_class", "STANDARD")
	} else {
		err = d.Set("storage_class", normalizeStorageClass(class))
	}
	mErr := multierror.Append(err,
		d.Set("size", object.Size),
		d.Set("etag", strings.Trim(object.ETag, `"`)),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return fmterr.Errorf("error setting OBS bucket attributes: %s", err)
	}

	return nil
}

func resourceObsBucketObjectDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := config.NewObjectStorageClient(config.GetRegion(d))
	if err != nil {
		return fmterr.Errorf("error creating OBS client: %s", err)
	}
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	input := obs.DeleteObjectInput{
		Bucket: bucket,
		Key:    key,
	}

	if _, ok := d.GetOk("version_id"); ok {
		// Bucket is versioned, we need to delete all versions
		vInput := obs.ListVersionsInput{
			Bucket: bucket,
		}
		out, err := client.ListVersions(&vInput)
		if err != nil {
			return fmterr.Errorf("failed listing OBS object versions: %s", err)
		}

		for _, v := range out.Versions {
			input.VersionId = v.VersionId
			_, err := client.DeleteObject(&input)
			if err != nil {
				return fmterr.Errorf("error deleting OBS object version of %s:\n%s,\n%s", key, v.VersionId, err)
			}
		}
		return nil
	}

	// Just delete the object
	_, err = client.DeleteObject(&input)
	if err != nil {
		return fmterr.Errorf("error deleting OBS bucket object: %s  Bucket: %q Object: %q", err, bucket, key)
	}

	return nil
}
