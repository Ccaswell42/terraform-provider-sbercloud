// ---------------------------------------------------------------
// *** AUTO GENERATED CODE ***
// @Product Organizations
// ---------------------------------------------------------------

package organizations

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccountCreate,
		UpdateContext: resourceAccountUpdate,
		ReadContext:   resourceAccountRead,
		DeleteContext: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		Description: "schema: Internal",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `Specifies the name of the account.`,
			},
			"parent_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the ID of the root or organization unit in which you want to create a new account.`,
			},
			"tags": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
				Description: `Specifies the key/value to attach to the account.`,
			},
			"urn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the uniform resource name of the account.`,
			},
			"joined_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates the time when the account was created.`,
			},
			"joined_method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `Indicates how an account joined an organization.`,
			},
		},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// createAccount: create Organizations account
	var (
		createAccountHttpUrl = "v1/organizations/accounts"
		createAccountProduct = "organizations"
	)
	createAccountClient, err := cfg.NewServiceClient(createAccountProduct, "")
	if err != nil {
		return diag.Errorf("error creating Organizations Client: %s", err)
	}

	createAccountPath := createAccountClient.Endpoint + createAccountHttpUrl

	createAccountOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	createAccountOpt.JSONBody = utils.RemoveNil(buildCreateAccountBodyParams(d))
	createAccountResp, err := createAccountClient.Request("POST", createAccountPath, &createAccountOpt)
	if err != nil {
		return diag.Errorf("error creating Account: %s", err)
	}

	createAccountRespBody, err := utils.FlattenResponse(createAccountResp)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := jmespath.Search("create_account_status.account_id", createAccountRespBody)
	if err != nil {
		return diag.Errorf("error creating Account: ID is not found in API response")
	}
	d.SetId(id.(string))

	stateId, err := jmespath.Search("create_account_status.id", createAccountRespBody)
	if err != nil {
		return diag.Errorf("error creating Account: state is not found in API response")
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"in_progress"},
		Target:       []string{"succeeded"},
		Refresh:      accountStateRefreshFunc(createAccountClient, stateId.(string)),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        10 * time.Second,
		PollInterval: 10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Organizations account (%s) to create: %s", id.(string), err)
	}

	if v, ok := d.GetOk("parent_id"); ok {
		parentID, err := getParentIdByAccountId(createAccountClient, id.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		err = moveAccount(createAccountClient, d.Id(), parentID, v.(string))
		if err != nil {
			return diag.Errorf("error moving Account %s to organization unit %s: %s", d.Id(), v.(string), err)
		}
	}

	return resourceAccountRead(ctx, d, meta)
}

func accountStateRefreshFunc(client *golangsdk.ServiceClient, accountStatusId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		getAccountStatusHttpUrl := "v1/organizations/create-account-status/{create_account_status_id}"
		getAccountStatusPath := client.Endpoint + getAccountStatusHttpUrl
		getAccountStatusPath = strings.ReplaceAll(getAccountStatusPath, "{create_account_status_id}", accountStatusId)

		getAccountStatusOpt := golangsdk.RequestOpts{
			KeepResponseBody: true,
		}
		getAccountStatusResp, err := client.Request("GET", getAccountStatusPath, &getAccountStatusOpt)
		if err != nil {
			return nil, "", err
		}

		getAccountStatusRespBody, err := utils.FlattenResponse(getAccountStatusResp)
		if err != nil {
			return nil, "", err
		}

		state, err := jmespath.Search("create_account_status.state", getAccountStatusRespBody)
		if err != nil {
			return nil, "", err
		}

		return getAccountStatusRespBody, state.(string), nil
	}
}

func buildCreateAccountBodyParams(d *schema.ResourceData) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"name": utils.ValueIngoreEmpty(d.Get("name")),
		"tags": utils.ExpandResourceTags(d.Get("tags").(map[string]interface{})),
	}
	return bodyParams
}

func resourceAccountRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	var mErr *multierror.Error

	// getAccount: Query Organizations account
	var (
		getAccountProduct = "organizations"
	)
	getAccountClient, err := cfg.NewServiceClient(getAccountProduct, "")
	if err != nil {
		return diag.Errorf("error creating Organizations Client: %s", err)
	}

	getAccountHttpUrl := "v1/organizations/accounts/{account_id}"
	getAccountPath := getAccountClient.Endpoint + getAccountHttpUrl
	getAccountPath = strings.ReplaceAll(getAccountPath, "{account_id}", d.Id())

	getAccountOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	getAccountResp, err := getAccountClient.Request("GET", getAccountPath, &getAccountOpt)

	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving Account")
	}

	getAccountRespBody, err := utils.FlattenResponse(getAccountResp)
	if err != nil {
		return diag.FromErr(err)
	}

	parentID, err := getParentIdByAccountId(getAccountClient, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	mErr = multierror.Append(
		mErr,
		d.Set("parent_id", parentID),
		d.Set("name", utils.PathSearch("account.name", getAccountRespBody, nil)),
		d.Set("urn", utils.PathSearch("account.urn", getAccountRespBody, nil)),
		d.Set("joined_at", utils.PathSearch("account.joined_at", getAccountRespBody, nil)),
		d.Set("joined_method", utils.PathSearch("account.join_method", getAccountRespBody, nil)),
	)

	tagMap, err := getTags(getAccountClient, accountsType, d.Id())
	if err != nil {
		log.Printf("[WARN] error fetching tags of Organizations account (%s): %s", d.Id(), err)
	} else {
		mErr = multierror.Append(mErr, d.Set("tags", tagMap))
	}

	return diag.FromErr(mErr.ErrorOrNil())
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)

	// updateAccount: update Organizations account
	var (
		updateAccountProduct = "organizations"
	)
	updateAccountClient, err := cfg.NewServiceClient(updateAccountProduct, "")
	if err != nil {
		return diag.Errorf("error creating Organizations Client: %s", err)
	}

	if d.HasChange("parent_id") {
		oldVal, newVal := d.GetChange("parent_id")
		err = moveAccount(updateAccountClient, d.Id(), oldVal.(string), newVal.(string))
		if err != nil {
			return diag.Errorf("error updating Account: %s", err)
		}
	}

	if d.HasChange("tags") {
		err = updateTags(d, updateAccountClient, accountsType, d.Id(), "tags")
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceAccountRead(ctx, d, meta)
}

func buildUpdateAccountBodyParams(oldOrganizationsUnitId, newOrganizationsUnitId string) map[string]interface{} {
	bodyParams := map[string]interface{}{
		"source_parent_id":      oldOrganizationsUnitId,
		"destination_parent_id": newOrganizationsUnitId,
	}
	return bodyParams
}

func moveAccount(client *golangsdk.ServiceClient, accountId, sourceParentID, destinationParentID string) error {
	// moveAccount: update Organizations account
	var (
		moveAccountHttpUrl = "v1/organizations/accounts/{account_id}/move"
	)
	moveAccountPath := client.Endpoint + moveAccountHttpUrl
	moveAccountPath = strings.ReplaceAll(moveAccountPath, "{account_id}", accountId)

	moveAccountOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	moveAccountOpt.JSONBody = utils.RemoveNil(buildUpdateAccountBodyParams(sourceParentID, destinationParentID))
	_, err := client.Request("POST", moveAccountPath, &moveAccountOpt)
	return err
}

func resourceAccountDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	errorMsg := "Deleting Organizations account is not supported. The account is only removed from the state," +
		" but it remains in the cloud."
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  errorMsg,
		},
	}
}

func getParentIdByAccountId(client *golangsdk.ServiceClient, accountID string) (string, error) {
	getParentHttpUrl := "v1/organizations/entities?child_id={account_id}"
	getParentPath := client.Endpoint + getParentHttpUrl
	getParentPath = strings.ReplaceAll(getParentPath, "{account_id}", accountID)

	getParentOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}
	getAccountResp, err := client.Request("GET", getParentPath, &getParentOpt)
	if err != nil {
		return "", fmt.Errorf("error retrieving parent by account_id: %s", accountID)
	}
	getAccountRespBody, err := utils.FlattenResponse(getAccountResp)
	if err != nil {
		return "", err
	}

	id := utils.PathSearch("entities|[0].id", getAccountRespBody, "").(string)

	return id, nil
}
