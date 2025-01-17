package huaweicloud

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/shares"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

const (
	prepaidUnitMonth int = 2
	prepaidUnitYear  int = 3

	autoRenewDisabled int = 0
	autoRenewEnabled  int = 1
)

func ResourceSFSTurbo() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSFSTurboCreate,
		ReadContext:   resourceSFSTurboRead,
		UpdateContext: resourceSFSTurboUpdate,
		DeleteContext: resourceSFSTurboDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(4, 64),
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"share_proto": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "NFS",
				ValidateFunc: validation.StringInSlice([]string{"NFS"}, false),
			},
			"share_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "STANDARD",
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"crypt_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"enhanced": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enterprise_project_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"dedicated_flavor": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"dedicated_storage_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags":          common.TagsSchema(),
			"charging_mode": common.SchemaChargingMode(nil),
			"period_unit":   common.SchemaPeriodUnit(nil),
			"period":        common.SchemaPeriod(nil),
			"auto_renew":    common.SchemaAutoRenewUpdatable(nil),
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"available_capacity": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func buildTurboCreateOpts(cfg *config.Config, d *schema.ResourceData) shares.CreateOpts {
	metaOpts := shares.Metadata{}
	if v, ok := d.GetOk("crypt_key_id"); ok {
		metaOpts.CryptKeyID = v.(string)
	}
	if _, ok := d.GetOk("enhanced"); ok {
		metaOpts.ExpandType = "bandwidth"
	}
	if v, ok := d.GetOk("dedicated_flavor"); ok {
		metaOpts.DedicatedFlavor = v.(string)
	}
	if v, ok := d.GetOk("dedicated_storage_id"); ok {
		metaOpts.DedicatedStorageID = v.(string)
	}

	result := shares.CreateOpts{
		Share: shares.Share{
			Name:                d.Get("name").(string),
			Size:                d.Get("size").(int),
			ShareProto:          d.Get("share_proto").(string),
			ShareType:           d.Get("share_type").(string),
			VpcID:               d.Get("vpc_id").(string),
			SubnetID:            d.Get("subnet_id").(string),
			SecurityGroupID:     d.Get("security_group_id").(string),
			AvailabilityZone:    d.Get("availability_zone").(string),
			EnterpriseProjectId: cfg.GetEnterpriseProjectID(d),
			Metadata:            metaOpts,
		},
	}
	if d.Get("charging_mode") == "prePaid" {
		billing := shares.BssParam{
			PeriodNum: d.Get("period").(int),
			IsAutoPay: utils.Int(1), // Always enable auto-pay.
		}
		if d.Get("period_unit").(string) == "month" {
			billing.PeriodType = prepaidUnitMonth
		} else {
			billing.PeriodType = prepaidUnitYear
		}
		if d.Get("auto_renew").(string) == "true" {
			billing.IsAutoRenew = utils.Int(autoRenewEnabled)
		} else {
			billing.IsAutoRenew = utils.Int(autoRenewDisabled)
		}
		result.BssParam = &billing
	}
	return result
}

func resourceSFSTurboCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	sfsClient, err := cfg.SfsV1Client(GetRegion(d, cfg))
	if err != nil {
		return diag.Errorf("error creating SFS v1 client: %s", err)
	}

	createOpts := buildTurboCreateOpts(cfg, d)
	logp.Printf("[DEBUG] create sfs turbo with option: %+v", createOpts)
	resp, err := shares.Create(sfsClient, createOpts).Extract()
	if err != nil {
		return diag.Errorf("error creating SFS Turbo: %s", err)
	}

	if d.Get("charging_mode").(string) == "prePaid" {
		orderId := resp.OrderId
		if orderId == "" {
			return diag.Errorf("unable to find the order ID, this is a COM (Cloud Order Management) error, " +
				"please contact service for help and check your order status on the console.")
		}
		bssClient, err := cfg.BssV2Client(GetRegion(d, cfg))
		if err != nil {
			return diag.Errorf("error creating BSS v2 client: %s", err)
		}
		err = common.WaitOrderComplete(ctx, bssClient, orderId, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
		resourceId, err := common.WaitOrderResourceComplete(ctx, bssClient, orderId, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(resourceId)
	} else {
		d.SetId(resp.ID)
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"100"},
		Target:       []string{"200"},
		Refresh:      waitForSFSTurboStatus(sfsClient, resp.ID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 3 * time.Second,
	}
	_, StateErr := stateConf.WaitForState()
	if StateErr != nil {
		return diag.Errorf("error waiting for SFS Turbo (%s) to become ready: %s ", d.Id(), StateErr)
	}

	// add tags
	if err := utils.CreateResourceTags(sfsClient, d, "sfs-turbo", d.Id()); err != nil {
		return diag.Errorf("error setting tags of SFS Turbo %s: %s", d.Id(), err)
	}

	return resourceSFSTurboRead(ctx, d, meta)
}

func resourceSFSTurboRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	sfsClient, err := config.SfsV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating SFS v1 client: %s", err)
	}

	n, err := shares.Get(sfsClient, d.Id()).Extract()
	if err != nil {
		return common.CheckDeletedDiag(d, err, "SFS Turbo")
	}

	d.Set("name", n.Name)
	d.Set("share_proto", n.ShareProto)
	d.Set("share_type", n.ShareType)
	d.Set("vpc_id", n.VpcID)
	d.Set("subnet_id", n.SubnetID)
	d.Set("security_group_id", n.SecurityGroupID)
	d.Set("version", n.Version)
	d.Set("region", GetRegion(d, config))
	d.Set("availability_zone", n.AvailabilityZone)
	d.Set("available_capacity", n.AvailCapacity)
	d.Set("export_location", n.ExportLocation)
	d.Set("crypt_key_id", n.CryptKeyID)
	d.Set("enterprise_project_id", n.EnterpriseProjectId)
	// Cannot obtain the billing parameters for pre-paid.

	// n.Size is a string of float64, should convert it to int
	if fsize, err := strconv.ParseFloat(n.Size, 64); err == nil {
		if err = d.Set("size", int(fsize)); err != nil {
			return diag.Errorf("error reading size of SFS Turbo: %s", err)
		}
	}

	if n.ExpandType == "bandwidth" {
		d.Set("enhanced", true)
	} else {
		d.Set("enhanced", false)
	}

	var status string
	if n.SubStatus != "" {
		status = n.SubStatus
	} else {
		status = n.Status
	}
	d.Set("status", status)

	// set tags
	err = utils.SetResourceTagsToState(d, sfsClient, "sfs-turbo", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func buildTurboUpdateOpts(newSize int, isPrePaid bool) shares.ExpandOpts {
	expandOpts := shares.ExtendOpts{
		NewSize: newSize,
	}
	if isPrePaid {
		expandOpts.BssParam = &shares.BssParamExtend{
			IsAutoPay: utils.Int(1),
		}
	}
	return shares.ExpandOpts{
		Extend: expandOpts,
	}
}

func resourceSFSTurboUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := GetRegion(d, cfg)
	sfsClient, err := cfg.SfsV1Client(region)
	if err != nil {
		return diag.Errorf("error creating SFS v1 client: %s", err)
	}

	resourceId := d.Id()
	if d.HasChange("size") {
		old, newSize := d.GetChange("size")
		if old.(int) > newSize.(int) {
			return diag.Errorf("shrinking SFS Turbo size is not supported")
		}

		isPrePaid := d.Get("charging_mode").(string) == "prePaid"
		updateOpts := buildTurboUpdateOpts(newSize.(int), isPrePaid)
		resp, err := shares.Expand(sfsClient, d.Id(), updateOpts).Extract()
		if err != nil {
			return diag.Errorf("error expanding SFS Turbo size: %s", err)
		}

		if isPrePaid {
			orderId := resp.OrderId
			if orderId == "" {
				return diag.Errorf("unable to find the order ID, this is a COM (Cloud Order Management) error, " +
					"please contact service for help and check your order status on the console.")
			}
			bssClient, err := cfg.BssV2Client(region)
			if err != nil {
				return diag.Errorf("error creating BSS v2 client: %s", err)
			}
			err = common.WaitOrderComplete(ctx, bssClient, orderId, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
			_, err = common.WaitOrderResourceComplete(ctx, bssClient, orderId, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return diag.FromErr(err)
			}
		}
		stateConf := &resource.StateChangeConf{
			Pending:      []string{"121"},
			Target:       []string{"221", "200"},
			Refresh:      waitForSFSTurboSubStatus(sfsClient, resourceId),
			Timeout:      d.Timeout(schema.TimeoutUpdate),
			PollInterval: 5 * time.Second,
		}
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.Errorf("Error updating HuaweiCloud SFS Turbo: %s", err)
		}
	}

	// update tags
	if d.HasChange("tags") {
		if err := updateSFSTurboTags(sfsClient, d); err != nil {
			return diag.Errorf("error updating tags of SFS Turbo %s: %s", resourceId, err)
		}
	}

	if d.HasChange("auto_renew") {
		bssClient, err := cfg.BssV2Client(region)
		if err != nil {
			return diag.Errorf("error creating BSS V2 client: %s", err)
		}
		if err = common.UpdateAutoRenew(bssClient, d.Get("auto_renew").(string), resourceId); err != nil {
			return diag.Errorf("error updating the auto-renew of the SFS Turbo (%s): %s", resourceId, err)
		}
	}

	return resourceSFSTurboRead(ctx, d, meta)
}

func updateSFSTurboTags(client *golangsdk.ServiceClient, d *schema.ResourceData) error {
	// remove old tags
	oldKeys := getOldTagKeys(d)
	if err := utils.DeleteResourceTagsWithKeys(client, oldKeys, "sfs-turbo", d.Id()); err != nil {
		return err
	}

	// set new tags
	return utils.CreateResourceTags(client, d, "sfs-turbo", d.Id())
}

func getOldTagKeys(d *schema.ResourceData) []string {
	oRaw, _ := d.GetChange("tags")
	var tagKeys []string
	if oMap := oRaw.(map[string]interface{}); len(oMap) > 0 {
		for k := range oMap {
			tagKeys = append(tagKeys, k)
		}
	}
	return tagKeys
}

func resourceSFSTurboDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	sfsClient, err := config.SfsV1Client(GetRegion(d, config))
	if err != nil {
		return diag.Errorf("error creating SFS v1 client: %s", err)
	}

	resourceId := d.Id()
	// for prePaid mode, we should unsubscribe the resource
	if d.Get("charging_mode").(string) == "prePaid" {
		err := common.UnsubscribePrePaidResource(d, config, []string{resourceId})
		if err != nil {
			return diag.Errorf("error unsubscribing SFS Turbo: %s", err)
		}
	} else {
		err = shares.Delete(sfsClient, resourceId).ExtractErr()
		if err != nil {
			return common.CheckDeletedDiag(d, err, "SFS Turbo")
		}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"100", "200"},
		Target:     []string{"deleted"},
		Refresh:    waitForSFSTurboStatus(sfsClient, resourceId),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error deleting SFS Turbo: %s", err)
	}
	return nil
}

func waitForSFSTurboStatus(sfsClient *golangsdk.ServiceClient, shareId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := shares.Get(sfsClient, shareId).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				logp.Printf("[INFO] Successfully deleted shared File %s", shareId)
				return r, "deleted", nil
			}
			return r, "error", err
		}

		return r, r.Status, nil
	}
}

func waitForSFSTurboSubStatus(sfsClient *golangsdk.ServiceClient, shareId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r, err := shares.Get(sfsClient, shareId).Extract()
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				logp.Printf("[INFO] Successfully deleted shared File %s", shareId)
				return r, "deleted", nil
			}
			return r, "error", err
		}

		var status string
		if r.SubStatus != "" {
			status = r.SubStatus
		} else {
			status = r.Status
		}
		return r, status, nil
	}
}
