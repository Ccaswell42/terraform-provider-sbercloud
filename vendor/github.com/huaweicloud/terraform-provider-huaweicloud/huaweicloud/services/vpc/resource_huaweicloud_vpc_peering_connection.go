package vpc

import (
	"context"
	"log"
	"time"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/networking/v2/peerings"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceVpcPeeringConnectionV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPeeringV2Create,
		ReadContext:   resourceVPCPeeringV2Read,
		UpdateContext: resourceVPCPeeringV2Update,
		DeleteContext: resourceVPCPeeringV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{ // request and response parameters
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: utils.ValidateString64WithChinese,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVPCPeeringV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	peeringClient, err := config.NetworkingV2Client(config.GetRegion(d))

	if err != nil {
		return diag.Errorf("error creating Vpc Peering Connection Client: %s", err)
	}

	requestvpcinfo := peerings.VpcInfo{
		VpcId: d.Get("vpc_id").(string),
	}

	acceptvpcinfo := peerings.VpcInfo{
		VpcId:    d.Get("peer_vpc_id").(string),
		TenantId: d.Get("peer_tenant_id").(string),
	}

	createOpts := peerings.CreateOpts{
		Name:           d.Get("name").(string),
		RequestVpcInfo: requestvpcinfo,
		AcceptVpcInfo:  acceptvpcinfo,
	}

	n, err := peerings.Create(peeringClient, createOpts).Extract()

	if err != nil {
		return diag.Errorf("error creating Vpc Peering Connection: %s", err)
	}

	log.Printf("[INFO] Vpc Peering Connection ID: %s", n.ID)

	log.Printf("[INFO] Waiting for Vpc Peering Connection(%s) to become available", n.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"PENDING_ACCEPTANCE", "ACTIVE"},
		Refresh:    waitForVpcPeeringActive(peeringClient, n.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		log.Printf("Error creating Vpc Peering Connection: %s", err)
	}
	d.SetId(n.ID)

	return resourceVPCPeeringV2Read(ctx, d, meta)

}

func resourceVPCPeeringV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	peeringClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating Vpc Peering Connection Client: %s", err)
	}

	n, err := peerings.Get(peeringClient, d.Id()).Extract()
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			d.SetId("")
			return nil
		}

		return diag.Errorf("error retrieving Vpc Peering Connection: %s", err)
	}

	d.Set("name", n.Name)
	d.Set("status", n.Status)
	d.Set("vpc_id", n.RequestVpcInfo.VpcId)
	d.Set("peer_vpc_id", n.AcceptVpcInfo.VpcId)
	d.Set("peer_tenant_id", n.AcceptVpcInfo.TenantId)
	d.Set("region", config.GetRegion(d))

	return nil
}

func resourceVPCPeeringV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	peeringClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating Vpc Peering Connection Client: %s", err)
	}

	var updateOpts peerings.UpdateOpts

	updateOpts.Name = d.Get("name").(string)

	_, err = peerings.Update(peeringClient, d.Id(), updateOpts).Extract()
	if err != nil {
		return diag.Errorf("error updating Vpc Peering Connection: %s", err)
	}

	return resourceVPCPeeringV2Read(ctx, d, meta)
}

func resourceVPCPeeringV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*config.Config)
	peeringClient, err := config.NetworkingV2Client(config.GetRegion(d))
	if err != nil {
		return diag.Errorf("error creating Vpc Peering Connection Client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    waitForVpcPeeringDelete(peeringClient, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error deleting Vpc Peering Connection: %s", err)
	}

	d.SetId("")
	return nil
}

func waitForVpcPeeringActive(peeringClient *golangsdk.ServiceClient, peeringId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		n, err := peerings.Get(peeringClient, peeringId).Extract()
		if err != nil {
			return nil, "", err
		}

		if n.Status == "PENDING_ACCEPTANCE" || n.Status == "ACTIVE" {
			return n, n.Status, nil
		}

		return n, "CREATING", nil
	}
}

func waitForVpcPeeringDelete(peeringClient *golangsdk.ServiceClient, peeringId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		r, err := peerings.Get(peeringClient, peeringId).Extract()

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[INFO] Successfully deleted vpc peering connection %s", peeringId)
				return r, "DELETED", nil
			}
			return r, "ACTIVE", err
		}

		err = peerings.Delete(peeringClient, peeringId).ExtractErr()

		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				log.Printf("[INFO] Successfully deleted vpc peering connection %s", peeringId)
				return r, "DELETED", nil
			}
			if errCode, ok := err.(golangsdk.ErrUnexpectedResponseCode); ok {
				if errCode.Actual == 409 {
					return r, "ACTIVE", nil
				}
			}
			return r, "ACTIVE", err
		}

		return r, "ACTIVE", nil
	}
}
