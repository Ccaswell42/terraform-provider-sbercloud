// ---------------------------------------------------------------
// *** AUTO GENERATED CODE ***
// @Product WAF
// ---------------------------------------------------------------

package waf

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmespath/go-jmespath"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func ResourceRulePreciseProtection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRulePreciseProtectionCreate,
		UpdateContext: resourceRulePreciseProtectionUpdate,
		ReadContext:   resourceRulePreciseProtectionRead,
		DeleteContext: resourceRulePreciseProtectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWAFRuleImportState,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `Specifies the policy ID of WAF precise protection rule.`,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the name of WAF precise protection rule.`,
			},
			"priority": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: `Specifies the priority of a rule.`,
			},
			"conditions": {
				Type:        schema.TypeList,
				Elem:        conditionsSchema(),
				Required:    true,
				Description: `Specifies the match condition list.`,
			},
			"enterprise_project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `Specifies the enterprise project ID of WAF precise protection rule.`,
			},
			"action": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "block",
				Description: `Specifies the protective action of the precise protection rule.`,
			},
			"status": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: `Specifies the status of WAF precise protection rule.`,
			},
			"start_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the time when the precise protection rule takes effect.`,
			},
			"end_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `Specifies the time when the precise protection rule expires.`,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the description of WAF precise protection rule.`,
			},
		},
	}
}

func conditionsSchema() *schema.Resource {
	sc := schema.Resource{
		Schema: map[string]*schema.Schema{
			"field": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the field of the condition.`,
			},
			"logic": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `Specifies the condition matching logic.`,
			},
			"subfield": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the subfield of the condition.`,
			},
			"content": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the content of the match condition.`,
			},
			"reference_table_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: `Specifies the reference table id.`,
			},
		},
	}
	return &sc
}

func resourceRulePreciseProtectionCreate(ctx context.Context, d *schema.ResourceData,
	meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var (
		preciseProtectionHttpUrl = "v1/{project_id}/waf/policy/{policy_id}/custom"
		preciseProtectionProduct = "waf"
	)
	preciseProtectionClient, err := cfg.NewServiceClient(preciseProtectionProduct, region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	createPath := preciseProtectionClient.Endpoint + preciseProtectionHttpUrl
	createPath = strings.ReplaceAll(createPath, "{project_id}", preciseProtectionClient.ProjectID)
	createPath = strings.ReplaceAll(createPath, "{policy_id}", fmt.Sprintf("%v", d.Get("policy_id")))
	createPath += buildQueryParams(d, cfg)
	createOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}

	bodyParam, err := buildCreateOrUpdateBodyParams(d)
	if err != nil {
		return diag.FromErr(err)
	}

	createOpt.JSONBody = utils.RemoveNil(bodyParam)
	createResp, err := preciseProtectionClient.Request("POST", createPath, &createOpt)
	if err != nil {
		return diag.Errorf("error creating RulePreciseProtection: %s", err)
	}

	createRespBody, err := utils.FlattenResponse(createResp)
	if err != nil {
		return diag.FromErr(err)
	}

	id, err := jmespath.Search("id", createRespBody)
	if err != nil {
		return diag.Errorf("error creating RulePreciseProtection: ID is not found in API response")
	}
	d.SetId(id.(string))

	if d.Get("status").(int) == 0 {
		if err := updateRuleStatus(preciseProtectionClient, d, cfg, "custom"); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceRulePreciseProtectionRead(ctx, d, meta)
}

func updateRuleStatus(client *golangsdk.ServiceClient, d *schema.ResourceData, cfg *config.Config,
	ruleType string) error {
	var (
		updateWAFRuleStatusHttpUrl = "v1/{project_id}/waf/policy/{policy_id}/{rule_type}/{rule_id}/status"
		policyID                   = fmt.Sprintf("%v", d.Get("policy_id"))
	)

	updateStatusPath := client.Endpoint + updateWAFRuleStatusHttpUrl
	updateStatusPath = strings.ReplaceAll(updateStatusPath, "{project_id}", client.ProjectID)
	updateStatusPath = strings.ReplaceAll(updateStatusPath, "{policy_id}", policyID)
	updateStatusPath = strings.ReplaceAll(updateStatusPath, "{rule_type}", ruleType)
	updateStatusPath = strings.ReplaceAll(updateStatusPath, "{rule_id}", d.Id())
	updateStatusPath += buildQueryParams(d, cfg)

	updateWAFRuleStatusOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	updateWAFRuleStatusOpt.JSONBody = utils.RemoveNil(buildUpdateWAFRuleStatusBodyParams(d))
	_, err := client.Request("PUT", updateStatusPath, &updateWAFRuleStatusOpt)
	if err != nil {
		return fmt.Errorf("error updating %s rule status: %s", ruleType, err)
	}
	return nil
}

func buildCreateOrUpdateBodyParams(d *schema.ResourceData) (map[string]interface{}, error) {
	bodyParams := map[string]interface{}{
		"name":        utils.ValueIngoreEmpty(d.Get("name")),
		"priority":    utils.ValueIngoreEmpty(d.Get("priority")),
		"conditions":  buildConditionBodyParam(d.Get("conditions")),
		"action":      buildActionBodyParam(d),
		"description": utils.ValueIngoreEmpty(d.Get("description")),
	}

	if v, ok := d.GetOk("start_time"); ok {
		stamp, err := utils.FormatUTCTimeStamp(v.(string))
		if err != nil {
			return nil, err
		}
		bodyParams["start"] = stamp
		bodyParams["time"] = true
	}

	if v, ok := d.GetOk("end_time"); ok {
		stamp, err := utils.FormatUTCTimeStamp(v.(string))
		if err != nil {
			return nil, err
		}
		bodyParams["terminal"] = stamp
		bodyParams["time"] = true
	}
	return bodyParams, nil
}

func buildActionBodyParam(d *schema.ResourceData) map[string]interface{} {
	if v, ok := d.GetOk("action"); ok {
		return map[string]interface{}{
			"category": v,
		}
	}
	return nil
}

func buildConditionBodyParam(rawParams interface{}) []map[string]interface{} {
	if rawArray, ok := rawParams.([]interface{}); ok {
		if len(rawArray) == 0 {
			return nil
		}

		rst := make([]map[string]interface{}, len(rawArray))
		for i, v := range rawArray {
			raw := v.(map[string]interface{})
			rst[i] = map[string]interface{}{
				"category":        utils.ValueIngoreEmpty(raw["field"]),
				"index":           utils.ValueIngoreEmpty(raw["subfield"]),
				"logic_operation": utils.ValueIngoreEmpty(raw["logic"]),
				"contents":        buildContentBodyParam(raw),
				"value_list_id":   utils.ValueIngoreEmpty(raw["reference_table_id"]),
			}
		}
		return rst
	}
	return nil
}

func buildContentBodyParam(raw map[string]interface{}) []string {
	var contents []string
	if content := utils.ValueIngoreEmpty(raw["content"]); content != nil {
		contents = append(contents, content.(string))
	}
	return contents
}

func buildUpdateWAFRuleStatusBodyParams(d *schema.ResourceData) map[string]interface{} {
	return map[string]interface{}{
		"status": d.Get("status").(int),
	}
}

func buildQueryParams(d *schema.ResourceData, cfg *config.Config) string {
	epsId := cfg.GetEnterpriseProjectID(d)
	if epsId == "" {
		return ""
	}
	return fmt.Sprintf("?enterprise_project_id=%s", epsId)
}

func resourceRulePreciseProtectionRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var mErr *multierror.Error

	var (
		preciseProtectionHttpUrl = "v1/{project_id}/waf/policy/{policy_id}/custom/{rule_id}"
		product                  = "waf"
	)
	preciseProtectionClient, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	getRulePath := preciseProtectionClient.Endpoint + preciseProtectionHttpUrl
	getRulePath = strings.ReplaceAll(getRulePath, "{project_id}", preciseProtectionClient.ProjectID)
	getRulePath = strings.ReplaceAll(getRulePath, "{policy_id}", fmt.Sprintf("%v", d.Get("policy_id")))
	getRulePath = strings.ReplaceAll(getRulePath, "{rule_id}", d.Id())
	getRulePath += buildQueryParams(d, cfg)

	getRuleOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	getRuleResp, err := preciseProtectionClient.Request("GET", getRulePath, &getRuleOpt)

	if err != nil {
		return common.CheckDeletedDiag(d, err, "error retrieving RulePreciseProtection")
	}

	getRuleRespBody, err := utils.FlattenResponse(getRuleResp)
	if err != nil {
		return diag.FromErr(err)
	}
	mErr = multierror.Append(
		mErr,
		d.Set("region", region),
		d.Set("name", utils.PathSearch("name", getRuleRespBody, nil)),
		d.Set("policy_id", utils.PathSearch("policyid", getRuleRespBody, nil)),
		d.Set("description", utils.PathSearch("description", getRuleRespBody, nil)),
		d.Set("priority", utils.PathSearch("priority", getRuleRespBody, nil)),
		d.Set("status", utils.PathSearch("status", getRuleRespBody, nil)),
		d.Set("conditions", flattenRulePreciseProtectionConditions(getRuleRespBody)),
		d.Set("action", flattenRulePreciseProtectionAction(getRuleRespBody)),
		d.Set("start_time", flattenRulePreciseProtectionTime(getRuleRespBody, "start")),
		d.Set("end_time", flattenRulePreciseProtectionTime(getRuleRespBody, "terminal")),
	)
	return diag.FromErr(mErr.ErrorOrNil())
}

func flattenRulePreciseProtectionTime(resp interface{}, field string) string {
	if resp == nil {
		return ""
	}
	timestamp := utils.PathSearch(field, resp, nil)
	if timestamp == nil {
		return ""
	}
	return utils.FormatTimeStampUTC(int64(timestamp.(float64)))
}

func flattenRulePreciseProtectionAction(resp interface{}) string {
	if resp == nil {
		return ""
	}
	customAction := utils.PathSearch("action", resp, nil)
	if customAction == nil {
		return ""
	}
	customActionMap := customAction.(map[string]interface{})
	if v, ok := customActionMap["category"]; ok {
		return v.(string)
	}
	return ""
}

func flattenRulePreciseProtectionConditions(resp interface{}) []interface{} {
	if resp == nil {
		return nil
	}
	curJson := utils.PathSearch("conditions", resp, make([]interface{}, 0))
	curArray := curJson.([]interface{})
	rst := make([]interface{}, 0, len(curArray))
	for _, v := range curArray {
		rst = append(rst, map[string]interface{}{
			"field":              utils.PathSearch("category", v, nil),
			"subfield":           utils.PathSearch("index", v, nil),
			"logic":              utils.PathSearch("logic_operation", v, nil),
			"content":            utils.PathSearch("contents|[0]", v, nil),
			"reference_table_id": utils.PathSearch("value_list_id", v, nil),
		})
	}
	return rst
}

func resourceRulePreciseProtectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	preciseProtectionClient, err := cfg.NewServiceClient("waf", region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	updateWAFRulePreciseProtectionChanges := []string{
		"name",
		"priority",
		"conditions",
		"action",
		"start_time",
		"end_time",
		"description",
	}

	if d.HasChanges(updateWAFRulePreciseProtectionChanges...) {
		updateHttpUrl := "v1/{project_id}/waf/policy/{policy_id}/custom/{rule_id}"

		updatePath := preciseProtectionClient.Endpoint + updateHttpUrl
		updatePath = strings.ReplaceAll(updatePath, "{project_id}", preciseProtectionClient.ProjectID)
		updatePath = strings.ReplaceAll(updatePath, "{policy_id}", d.Get("policy_id").(string))
		updatePath = strings.ReplaceAll(updatePath, "{rule_id}", d.Id())
		updatePath += buildQueryParams(d, cfg)

		updateOpt := golangsdk.RequestOpts{
			MoreHeaders: map[string]string{
				"Content-Type": "application/json;charset=utf8",
			},
			KeepResponseBody: true,
			OkCodes: []int{
				200,
			},
		}
		bodyParam, err := buildCreateOrUpdateBodyParams(d)
		if err != nil {
			return diag.FromErr(err)
		}

		updateOpt.JSONBody = utils.RemoveNil(bodyParam)
		_, err = preciseProtectionClient.Request("PUT", updatePath, &updateOpt)
		if err != nil {
			return diag.Errorf("error updating RulePreciseProtection: %s", err)
		}
	}

	if d.HasChange("status") {
		if err := updateRuleStatus(preciseProtectionClient, d, cfg, "custom"); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceRulePreciseProtectionRead(ctx, d, meta)
}

func resourceRulePreciseProtectionDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	region := cfg.GetRegion(d)

	var (
		deleteHttpUrl = "v1/{project_id}/waf/policy/{policy_id}/custom/{rule_id}"
		product       = "waf"
	)
	preciseProtectionClient, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return diag.Errorf("error creating WAF Client: %s", err)
	}

	deletePath := preciseProtectionClient.Endpoint + deleteHttpUrl
	deletePath = strings.ReplaceAll(deletePath, "{project_id}", preciseProtectionClient.ProjectID)
	deletePath = strings.ReplaceAll(deletePath, "{policy_id}", d.Get("policy_id").(string))
	deletePath = strings.ReplaceAll(deletePath, "{rule_id}", d.Id())
	deletePath += buildQueryParams(d, cfg)

	deleteOpt := golangsdk.RequestOpts{
		MoreHeaders: map[string]string{
			"Content-Type": "application/json;charset=utf8",
		},
		KeepResponseBody: true,
		OkCodes: []int{
			200,
		},
	}
	_, err = preciseProtectionClient.Request("DELETE", deletePath, &deleteOpt)
	if err != nil {
		return diag.Errorf("error deleting RulePreciseProtection: %s", err)
	}
	return nil
}

func resourceWAFRuleImportState(_ context.Context, d *schema.ResourceData,
	_ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	partLength := len(parts)

	if partLength == 3 {
		d.SetId(parts[1])
		mErr := multierror.Append(nil,
			d.Set("policy_id", parts[0]),
			d.Set("enterprise_project_id", parts[2]),
		)
		if err := mErr.ErrorOrNil(); err != nil {
			return nil, fmt.Errorf("failed to set value to state when import with epsid, %s", err)
		}
		return []*schema.ResourceData{d}, nil
	}
	if partLength == 2 {
		d.SetId(parts[1])
		mErr := multierror.Append(nil,
			d.Set("policy_id", parts[0]),
		)
		if err := mErr.ErrorOrNil(); err != nil {
			return nil, fmt.Errorf("failed to set value to state when import without epsid, %s", err)
		}
		return []*schema.ResourceData{d}, nil
	}
	return nil, fmt.Errorf("invalid format specified for import id," +
		" must be <policy_id>/<rule_id>/<enterprise_project_id> or <policy_id>/<rule_id>")
}
