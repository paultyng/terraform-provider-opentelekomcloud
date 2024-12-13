package dms

import (
	"context"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/topics"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDmsTopicsV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDmsTopicsV2Create,
		ReadContext:   resourceDmsTopicsV2Read,
		DeleteContext: resourceDmsTopicsV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateName,
			},
			"partition": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 200),
			},
			"replication": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 3),
			},
			"sync_replication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"retention_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 720),
			},
			"sync_message_flush": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"remain_partitions": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"max_partitions": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceDmsTopicsV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	DmsV2Client, err := common.ClientFromCtx(ctx, errCreationClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	createOpts := topics.CreateOpts{
		Name:             d.Get("name").(string),
		Partition:        d.Get("partition").(int),
		Replication:      d.Get("replication").(int),
		SyncReplication:  d.Get("sync_replication").(bool),
		RetentionTime:    d.Get("retention_time").(int),
		SyncMessageFlush: d.Get("sync_message_flush").(bool),
	}
	log.Printf("[DEBUG] Create Options: %#v", createOpts)
	v, err := topics.Create(DmsV2Client, d.Get("instance_id").(string), createOpts)
	if err != nil {
		return fmterr.Errorf("error creating OpenTelekomCloud queue: %s", err)
	}

	d.SetId(v.Name)

	clientCtx := common.CtxWithClient(ctx, DmsV2Client, errCreationClientV2)
	return resourceDmsTopicsV2Read(clientCtx, d, meta)
}

func resourceDmsTopicsV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)

	DmsV2Client, err := common.ClientFromCtx(ctx, errCreationClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	instanceId := d.Get("instance_id").(string)

	v, err := topics.List(DmsV2Client, instanceId)
	if err != nil {
		return common.CheckDeletedDiag(d, err, "DMS topic")
	}

	var fTopic topics.Topic
	found := false

	for _, topic := range v.Topics {
		if topic.Name == d.Id() {
			fTopic = topic
			found = true
			break
		}
	}
	if !found {
		return fmterr.Errorf("Provided topic doesn't exist")
	}

	var mErr *multierror.Error

	mErr = multierror.Append(
		mErr,
		d.Set("name", fTopic.Name),
		d.Set("partition", fTopic.Partition),
		d.Set("replication", fTopic.Replication),
		d.Set("sync_replication", fTopic.SyncReplication),
		d.Set("retention_time", fTopic.RetentionTime),
		d.Set("sync_message_flush", fTopic.SyncMessageFlush),
		d.Set("size", v.Size),
		d.Set("remain_partitions", v.RemainPartitions),
		d.Set("max_partitions", v.MaxPartitions),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDmsTopicsV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	DmsV2Client, err := common.ClientFromCtx(ctx, errCreationClientV2, func() (*golangsdk.ServiceClient, error) {
		return config.DmsV2Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationClientV2, err)
	}

	deleteOpts := []string{d.Id()}
	_, err = topics.Delete(DmsV2Client, d.Get("instance_id").(string), deleteOpts)
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud topic: %s", err)
	}

	log.Printf("[DEBUG] Dms topic %s deactivated.", d.Id())
	d.SetId("")
	return nil
}
