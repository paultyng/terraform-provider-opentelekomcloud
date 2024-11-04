package ddm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"

	"github.com/opentelekomcloud/gophertelekomcloud/openstack/ddm/v1/schemas"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/fmterr"
)

func ResourceDdmSchemaV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDdmSchemaV1Create,
		ReadContext:   resourceDdmSchemaV1Read,
		DeleteContext: resourceDdmSchemaV1Delete,

		Importer: &schema.ResourceImporter{
			StateContext: common.ImportByPath("ddm_instance_id", "name"),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateDDMSchemaName,
			},
			"ddm_instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: common.ValidateDDMSchemaName,
			},
			"shard_mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"cluster", "single",
				}, true),
			},
			"shard_number": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"shard_unit": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"rds": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 12,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"admin_username": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"admin_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"purge_rds_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"data_vips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"databases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_slot": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"updated": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rds_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"used_rds": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDdmSchemaV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	schemaDatabaseDetail := schemas.CreateDatabaseDetail{
		Name:        d.Get("name").(string),
		ShardMode:   d.Get("shard_mode").(string),
		ShardNumber: d.Get("shard_number").(int),
		ShardUnit:   d.Get("shard_unit").(int),
		UsedRds:     resourceDDMSchemaRDSV1(d),
	}

	createOpts := schemas.CreateSchemaOpts{
		Databases: []schemas.CreateDatabaseDetail{schemaDatabaseDetail},
	}

	ddmInstanceId := d.Get("ddm_instance_id").(string)
	ddmSchema, err := schemas.CreateSchema(client, ddmInstanceId, createOpts)
	if err != nil {
		return fmterr.Errorf("error getting OpenTelekomCloud DDM instance from result: %w", err)
	}
	schemaName := ddmSchema.Databases[0].Name
	id := fmt.Sprintf("%s/%s", ddmInstanceId, schemaName)
	d.SetId(id)
	log.Printf("[DEBUG] Create DDM schema %s", schemaName)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATE", "CREATING"},
		Target:     []string{"RUNNING"},
		Refresh:    schemaStateRefreshFunc(client, ddmInstanceId, schemaName),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      15 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmterr.Errorf("error waiting for OpenTelekomCloud DDM schema (%s) to become ready: %w", schemaName, err)
	}

	clientCtx := common.CtxWithClient(ctx, client, keyClientV1)
	return resourceDdmSchemaV1Read(clientCtx, d, meta)
}

func resourceDdmSchemaV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	schemaName := d.Get("name").(string)
	ddmInstanceId := d.Get("ddm_instance_id").(string)
	schema, err := schemas.QuerySchemaDetails(client, ddmInstanceId, schemaName)
	if err != nil {
		return fmterr.Errorf("error fetching DDM instance: %w", err)
	}

	log.Printf("[DEBUG] Retrieved instance %s: %#v", schemaName, schema.Database)

	mErr := multierror.Append(nil,
		d.Set("region", config.GetRegion(d)),
		d.Set("name", schema.Database.Name),
		d.Set("status", schema.Database.Status),
		d.Set("shard_mode", schema.Database.ShardMode),
		d.Set("shard_number", schema.Database.ShardNumber),
		d.Set("shard_unit", schema.Database.ShardUnit),
		d.Set("created_at", schema.Database.Created),
		d.Set("updated_at", schema.Database.Updated),
		d.Set("data_vips", schema.Database.DataVips),
	)

	var databasesList []map[string]interface{}
	for _, databaseRaw := range schema.Database.Databases {
		database := make(map[string]interface{})
		database["db_slot"] = databaseRaw.DbSlot
		database["name"] = databaseRaw.Name
		database["status"] = databaseRaw.Status
		database["created"] = databaseRaw.Created
		database["updated"] = databaseRaw.Updated
		database["id"] = databaseRaw.Id
		database["rds_name"] = databaseRaw.IdName
		databasesList = append(databasesList, database)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("databases", databasesList),
	)

	var rdsList []map[string]interface{}
	for _, rdsRaw := range schema.Database.UsedRds {
		rds := make(map[string]interface{})
		rds["id"] = rdsRaw.ID
		rds["name"] = rdsRaw.Name
		rds["status"] = rdsRaw.Status
		rdsList = append(rdsList, rds)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("used_rds", rdsList),
	)

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceDdmSchemaV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*cfg.Config)
	client, err := common.ClientFromCtx(ctx, keyClientV1, func() (*golangsdk.ServiceClient, error) {
		return config.DdmV1Client(config.GetRegion(d))
	})
	if err != nil {
		return fmterr.Errorf(errCreationV1Client, err)
	}

	log.Printf("[DEBUG] Deleting OpenTelekomCloud DDM schema %s", d.Id())

	schemaName := d.Get("name").(string)
	ddmInstanceId := d.Get("ddm_instance_id").(string)
	deleteRdsData := d.Get("purge_rds_on_delete").(bool)

	_, err = schemas.DeleteSchema(client, ddmInstanceId, schemaName, deleteRdsData)
	if err != nil {
		return fmterr.Errorf("error deleting OpenTelekomCloud DDM schema: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceDDMSchemaRDSV1(d *schema.ResourceData) []schemas.DatabaseInstancesParam {
	var rdsInstances []schemas.DatabaseInstancesParam

	rdsInputs := d.Get("rds").([]interface{})
	for i := range rdsInputs {
		rdsInput := rdsInputs[i].(map[string]interface{})
		rdsInstance := schemas.DatabaseInstancesParam{
			Id:            rdsInput["id"].(string),
			AdminUser:     rdsInput["admin_username"].(string),
			AdminPassword: rdsInput["admin_password"].(string),
		}

		rdsInstances = append(rdsInstances, rdsInstance)
	}
	return rdsInstances
}
