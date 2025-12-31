package rules

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ValidateRulesetConditions validates that conditions are appropriate for the ruleset.
// Returns errors for invalid combinations that would be silently dropped by the API.
//
// Critical Finding (tested 2025-12-22 against CheckMK 2.3.0p41):
// The CheckMK API silently drops certain conditions instead of rejecting them:
//   - host_label_rules ruleset: API silently drops host_labels and host_label_groups conditions
//   - service_label_rules ruleset: API silently drops service_labels and service_label_groups conditions
//
// This would cause infinite Terraform drift if not validated client-side:
//  1. User specifies label conditions on label ruleset
//  2. Terraform apply succeeds (API returns 200)
//  3. Terraform plan shows DRIFT (API returned empty conditions)
//  4. User re-applies... cycle repeats forever
func ValidateRulesetConditions(ruleset string, conditions map[string]interface{}) []string {
	var errors []string

	// host_label_rules cannot use host label conditions (circular dependency)
	// But it CAN use service label conditions (not circular)
	if isHostLabelRuleset(ruleset) && hasHostLabelConditions(conditions) {
		errors = append(errors,
			fmt.Sprintf("ruleset %q cannot use host_labels or host_label_groups conditions (circular dependency - would be silently dropped by API)", ruleset))
	}

	// service_label_rules cannot use service label conditions (circular dependency)
	// But it CAN use host label conditions (not circular)
	if isServiceLabelRuleset(ruleset) && hasServiceLabelConditions(conditions) {
		errors = append(errors,
			fmt.Sprintf("ruleset %q cannot use service_labels or service_label_groups conditions (circular dependency - would be silently dropped by API)", ruleset))
	}

	return errors
}

// isHostLabelRuleset returns true if the ruleset assigns host labels.
// Such rulesets cannot use host label conditions (circular dependency).
func isHostLabelRuleset(ruleset string) bool {
	return ruleset == "host_label_rules"
}

// isServiceLabelRuleset returns true if the ruleset assigns service labels.
// Such rulesets cannot use service label conditions (circular dependency).
func isServiceLabelRuleset(ruleset string) bool {
	return ruleset == "service_label_rules"
}

// isLabelRuleset returns true if the ruleset is any label ruleset.
// Kept for backwards compatibility and general checks.
func isLabelRuleset(ruleset string) bool {
	return isHostLabelRuleset(ruleset) || isServiceLabelRuleset(ruleset)
}

// hasHostLabelConditions checks if conditions contain host label filters.
// These conditions would be problematic on label rulesets.
func hasHostLabelConditions(conditions map[string]interface{}) bool {
	if conditions == nil {
		return false
	}

	// Check for host_labels condition
	if val, ok := conditions["host_labels"]; ok {
		// Check if it's non-empty
		switch v := val.(type) {
		case []interface{}:
			if len(v) > 0 {
				return true
			}
		case map[string]interface{}:
			if len(v) > 0 {
				return true
			}
		}
	}

	// Check for host_label_groups condition (CheckMK 2.3+)
	if val, ok := conditions["host_label_groups"]; ok {
		switch v := val.(type) {
		case []interface{}:
			if len(v) > 0 {
				return true
			}
		case map[string]interface{}:
			if len(v) > 0 {
				return true
			}
		}
	}

	return false
}

// hasServiceLabelConditions checks if conditions contain service label filters.
// These conditions would be problematic on service label rulesets.
func hasServiceLabelConditions(conditions map[string]interface{}) bool {
	if conditions == nil {
		return false
	}

	// Check for service_labels condition
	if val, ok := conditions["service_labels"]; ok {
		switch v := val.(type) {
		case []interface{}:
			if len(v) > 0 {
				return true
			}
		case map[string]interface{}:
			if len(v) > 0 {
				return true
			}
		}
	}

	// Check for service_label_groups condition (CheckMK 2.3+)
	if val, ok := conditions["service_label_groups"]; ok {
		switch v := val.(type) {
		case []interface{}:
			if len(v) > 0 {
				return true
			}
		case map[string]interface{}:
			if len(v) > 0 {
				return true
			}
		}
	}

	return false
}

// RuleConditionsSchema returns the Terraform schema for rule conditions.
// This schema can be reused across all rule resources.
//
// Based on CheckMK OpenAPI InputConditions schema (generic rules).
// Notification rules use RuleConditions which extends this with additional fields.
func RuleConditionsSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Conditions that determine when this rule applies. All specified conditions must match (AND logic).",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"host_name": schema.SingleNestedAttribute{
				MarkdownDescription: "Match hosts by name pattern.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"match_on": schema.ListAttribute{
						MarkdownDescription: "List of host name patterns (supports wildcards: `*`, `?`). Examples: `web-*`, `db-?.example.com`.",
						Required:            true,
						ElementType:         types.StringType,
					},
					"operator": schema.StringAttribute{
						MarkdownDescription: "Match operator. Options: `one_of` (match any pattern), `none_of` (match none).",
						Required:            true,
					},
				},
			},
			"host_tags": schema.ListNestedAttribute{
				MarkdownDescription: "Match hosts by tag conditions. Multiple tag conditions are combined with AND logic.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "Tag group key (e.g., `criticality`, `location`).",
							Required:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Tag match operator. Options: `is` (exact match), `is_not` (negation).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Tag value to match (e.g., `prod`, `us-east-1`).",
							Required:            true,
						},
					},
				},
			},
			"host_labels": schema.ListNestedAttribute{
				MarkdownDescription: "Match hosts by label conditions (CheckMK 2.2+). " +
					"**WARNING:** Cannot be used with `host_label_rules` or `service_label_rules` rulesets (circular dependency).",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "Label key (e.g., `environment`, `team`).",
							Required:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Label match operator. Options: `is` (exact match), `is_not` (negation).",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Label value to match (e.g., `production`, `platform`).",
							Required:            true,
						},
					},
				},
			},
			"host_label_groups": schema.ListAttribute{
				MarkdownDescription: "Match hosts by label group names (CheckMK 2.3+). " +
					"**WARNING:** Cannot be used with `host_label_rules` or `service_label_rules` rulesets (circular dependency).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"service_description": schema.SingleNestedAttribute{
				MarkdownDescription: "Match services by description pattern.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"match_on": schema.ListAttribute{
						MarkdownDescription: "List of service description patterns (supports wildcards). Examples: `CPU.*`, `Disk /var.*`.",
						Required:            true,
						ElementType:         types.StringType,
					},
					"operator": schema.StringAttribute{
						MarkdownDescription: "Match operator. Options: `one_of` (match any pattern), `none_of` (match none).",
						Required:            true,
					},
				},
			},
			"service_labels": schema.ListNestedAttribute{
				MarkdownDescription: "Match services by label conditions (CheckMK 2.2+). " +
					"**WARNING:** Cannot be used with `service_label_rules` ruleset (circular dependency).",
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "Label key.",
							Required:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "Label match operator. Options: `is`, `is_not`.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Label value to match.",
							Required:            true,
						},
					},
				},
			},
			"service_label_groups": schema.ListAttribute{
				MarkdownDescription: "Match services by label group names (CheckMK 2.3+). " +
					"**WARNING:** Cannot be used with `service_label_rules` ruleset (circular dependency).",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// ConvertConditionsToClient converts Terraform conditions to client format for API calls.
// This function extracts conditions from the Terraform types.Object and converts them
// to a map[string]interface{} that can be sent to the CheckMK API.
func ConvertConditionsToClient(ctx context.Context, conditions types.Object) (map[string]interface{}, error) {
	if conditions.IsNull() || conditions.IsUnknown() {
		// API requires conditions to be present (empty object), not null
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})
	attrs := conditions.Attributes()

	// Convert host_name condition
	if hostNameVal, ok := attrs["host_name"]; ok && !hostNameVal.IsNull() && !hostNameVal.IsUnknown() {
		hostNameObj, ok := hostNameVal.(types.Object)
		if ok && !hostNameObj.IsNull() {
			hostNameAttrs := hostNameObj.Attributes()
			matchOn := extractStringListFromAttr(hostNameAttrs["match_on"])
			operator := ""
			if opVal, ok := hostNameAttrs["operator"].(types.String); ok {
				operator = opVal.ValueString()
			}
			if len(matchOn) > 0 && operator != "" {
				result["host_name"] = map[string]interface{}{
					"match_on": matchOn,
					"operator": operator,
				}
			}
		}
	}

	// Convert host_tags condition
	if hostTagsVal, ok := attrs["host_tags"]; ok && !hostTagsVal.IsNull() && !hostTagsVal.IsUnknown() {
		hostTagsList, ok := hostTagsVal.(types.List)
		if ok && !hostTagsList.IsNull() {
			tags := extractLabelConditions(hostTagsList)
			if len(tags) > 0 {
				result["host_tags"] = tags
			}
		}
	}

	// Convert host_labels condition
	if hostLabelsVal, ok := attrs["host_labels"]; ok && !hostLabelsVal.IsNull() && !hostLabelsVal.IsUnknown() {
		hostLabelsList, ok := hostLabelsVal.(types.List)
		if ok && !hostLabelsList.IsNull() {
			labels := extractLabelConditions(hostLabelsList)
			if len(labels) > 0 {
				result["host_labels"] = labels
			}
		}
	}

	// Convert host_label_groups condition
	if hostLabelGroupsVal, ok := attrs["host_label_groups"]; ok && !hostLabelGroupsVal.IsNull() && !hostLabelGroupsVal.IsUnknown() {
		groups := extractStringListFromAttr(hostLabelGroupsVal)
		if len(groups) > 0 {
			result["host_label_groups"] = groups
		}
	}

	// Convert service_description condition
	if svcDescVal, ok := attrs["service_description"]; ok && !svcDescVal.IsNull() && !svcDescVal.IsUnknown() {
		svcDescObj, ok := svcDescVal.(types.Object)
		if ok && !svcDescObj.IsNull() {
			svcDescAttrs := svcDescObj.Attributes()
			matchOn := extractStringListFromAttr(svcDescAttrs["match_on"])
			operator := ""
			if opVal, ok := svcDescAttrs["operator"].(types.String); ok {
				operator = opVal.ValueString()
			}
			if len(matchOn) > 0 && operator != "" {
				result["service_description"] = map[string]interface{}{
					"match_on": matchOn,
					"operator": operator,
				}
			}
		}
	}

	// Convert service_labels condition
	if svcLabelsVal, ok := attrs["service_labels"]; ok && !svcLabelsVal.IsNull() && !svcLabelsVal.IsUnknown() {
		svcLabelsList, ok := svcLabelsVal.(types.List)
		if ok && !svcLabelsList.IsNull() {
			labels := extractLabelConditions(svcLabelsList)
			if len(labels) > 0 {
				result["service_labels"] = labels
			}
		}
	}

	// Convert service_label_groups condition
	if svcLabelGroupsVal, ok := attrs["service_label_groups"]; ok && !svcLabelGroupsVal.IsNull() && !svcLabelGroupsVal.IsUnknown() {
		groups := extractStringListFromAttr(svcLabelGroupsVal)
		if len(groups) > 0 {
			result["service_label_groups"] = groups
		}
	}

	return result, nil
}

// extractStringListFromAttr extracts a []string from an attr.Value that should be a types.List of strings.
func extractStringListFromAttr(val attr.Value) []string {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil
	}

	list, ok := val.(types.List)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(list.Elements()))
	for _, elem := range list.Elements() {
		if strVal, ok := elem.(types.String); ok && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

// extractLabelConditions extracts label/tag conditions from a types.List.
// Each element is an object with key, operator, value attributes.
func extractLabelConditions(list types.List) []map[string]interface{} {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	result := make([]map[string]interface{}, 0, len(elements))

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok || obj.IsNull() {
			continue
		}

		attrs := obj.Attributes()
		condition := make(map[string]interface{})

		if keyVal, ok := attrs["key"].(types.String); ok && !keyVal.IsNull() {
			condition["key"] = keyVal.ValueString()
		}
		if opVal, ok := attrs["operator"].(types.String); ok && !opVal.IsNull() {
			condition["operator"] = opVal.ValueString()
		}
		if valVal, ok := attrs["value"].(types.String); ok && !valVal.IsNull() {
			condition["value"] = valVal.ValueString()
		}

		// Only add if all required fields are present
		if condition["key"] != nil && condition["operator"] != nil && condition["value"] != nil {
			result = append(result, condition)
		}
	}

	return result
}

// ConvertConditionsFromClient converts client conditions to Terraform format.
// This function takes conditions from the CheckMK API response and converts them
// to a types.Object that can be stored in Terraform state.
func ConvertConditionsFromClient(ctx context.Context, conditions map[string]interface{}) (types.Object, error) {
	attrTypes := RuleConditionsAttrTypes()

	// If conditions is empty, return a null object with the proper types
	if len(conditions) == 0 {
		return types.ObjectNull(attrTypes), nil
	}

	// Build attribute values
	attrValues := make(map[string]attr.Value)

	// Convert host_name
	if hostNameRaw, ok := conditions["host_name"]; ok && hostNameRaw != nil {
		hostNameObj, err := convertMatchConditionFromClient(hostNameRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting host_name: %w", err)
		}
		attrValues["host_name"] = hostNameObj
	} else {
		attrValues["host_name"] = types.ObjectNull(matchConditionAttrTypes())
	}

	// Convert host_tags
	if hostTagsRaw, ok := conditions["host_tags"]; ok && hostTagsRaw != nil {
		hostTagsList, err := convertLabelConditionsFromClient(hostTagsRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting host_tags: %w", err)
		}
		attrValues["host_tags"] = hostTagsList
	} else {
		attrValues["host_tags"] = types.ListNull(labelConditionObjectType())
	}

	// Convert host_labels
	if hostLabelsRaw, ok := conditions["host_labels"]; ok && hostLabelsRaw != nil {
		hostLabelsList, err := convertLabelConditionsFromClient(hostLabelsRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting host_labels: %w", err)
		}
		attrValues["host_labels"] = hostLabelsList
	} else {
		attrValues["host_labels"] = types.ListNull(labelConditionObjectType())
	}

	// Convert host_label_groups
	if hostLabelGroupsRaw, ok := conditions["host_label_groups"]; ok && hostLabelGroupsRaw != nil {
		hostLabelGroupsList, err := convertStringListFromClient(hostLabelGroupsRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting host_label_groups: %w", err)
		}
		attrValues["host_label_groups"] = hostLabelGroupsList
	} else {
		attrValues["host_label_groups"] = types.ListNull(types.StringType)
	}

	// Convert service_description
	if svcDescRaw, ok := conditions["service_description"]; ok && svcDescRaw != nil {
		svcDescObj, err := convertMatchConditionFromClient(svcDescRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting service_description: %w", err)
		}
		attrValues["service_description"] = svcDescObj
	} else {
		attrValues["service_description"] = types.ObjectNull(matchConditionAttrTypes())
	}

	// Convert service_labels
	if svcLabelsRaw, ok := conditions["service_labels"]; ok && svcLabelsRaw != nil {
		svcLabelsList, err := convertLabelConditionsFromClient(svcLabelsRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting service_labels: %w", err)
		}
		attrValues["service_labels"] = svcLabelsList
	} else {
		attrValues["service_labels"] = types.ListNull(labelConditionObjectType())
	}

	// Convert service_label_groups
	if svcLabelGroupsRaw, ok := conditions["service_label_groups"]; ok && svcLabelGroupsRaw != nil {
		svcLabelGroupsList, err := convertStringListFromClient(svcLabelGroupsRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting service_label_groups: %w", err)
		}
		attrValues["service_label_groups"] = svcLabelGroupsList
	} else {
		attrValues["service_label_groups"] = types.ListNull(types.StringType)
	}

	obj, diags := types.ObjectValue(attrTypes, attrValues)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), fmt.Errorf("creating conditions object: %s", diags.Errors()[0].Detail())
	}
	return obj, nil
}

// matchConditionAttrTypes returns the attribute types for a match condition object.
func matchConditionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"match_on": types.ListType{ElemType: types.StringType},
		"operator": types.StringType,
	}
}

// labelConditionAttrTypes returns the attribute types for a label/tag condition object.
func labelConditionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":      types.StringType,
		"operator": types.StringType,
		"value":    types.StringType,
	}
}

// labelConditionObjectType returns the object type for a label/tag condition.
func labelConditionObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: labelConditionAttrTypes()}
}

// convertMatchConditionFromClient converts a match condition (host_name or service_description)
// from API format to Terraform types.Object.
func convertMatchConditionFromClient(raw interface{}) (types.Object, error) {
	attrTypes := matchConditionAttrTypes()

	if raw == nil {
		return types.ObjectNull(attrTypes), nil
	}

	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return types.ObjectNull(attrTypes), fmt.Errorf("expected map, got %T", raw)
	}

	if len(rawMap) == 0 {
		return types.ObjectNull(attrTypes), nil
	}

	// Extract match_on
	var matchOnList types.List
	if matchOnRaw, ok := rawMap["match_on"]; ok && matchOnRaw != nil {
		matchOnStrList, err := convertStringListFromClient(matchOnRaw)
		if err != nil {
			return types.ObjectNull(attrTypes), fmt.Errorf("converting match_on: %w", err)
		}
		matchOnList = matchOnStrList
	} else {
		matchOnList = types.ListNull(types.StringType)
	}

	// Extract operator
	var operatorVal types.String
	if opRaw, ok := rawMap["operator"]; ok && opRaw != nil {
		if opStr, ok := opRaw.(string); ok {
			operatorVal = types.StringValue(opStr)
		} else {
			operatorVal = types.StringNull()
		}
	} else {
		operatorVal = types.StringNull()
	}

	attrValues := map[string]attr.Value{
		"match_on": matchOnList,
		"operator": operatorVal,
	}

	obj, diags := types.ObjectValue(attrTypes, attrValues)
	if diags.HasError() {
		return types.ObjectNull(attrTypes), fmt.Errorf("creating match condition object: %s", diags.Errors()[0].Detail())
	}
	return obj, nil
}

// convertLabelConditionsFromClient converts a list of label/tag conditions from API format
// to Terraform types.List.
func convertLabelConditionsFromClient(raw interface{}) (types.List, error) {
	elemType := labelConditionObjectType()

	if raw == nil {
		return types.ListNull(elemType), nil
	}

	rawList, ok := raw.([]interface{})
	if !ok {
		return types.ListNull(elemType), fmt.Errorf("expected list, got %T", raw)
	}

	// Return null for empty arrays to avoid plan/state mismatches
	// (API returns [] but Terraform config doesn't specify the field)
	if len(rawList) == 0 {
		return types.ListNull(elemType), nil
	}

	elements := make([]attr.Value, 0, len(rawList))
	for _, item := range rawList {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		attrValues := make(map[string]attr.Value)

		if keyRaw, ok := itemMap["key"]; ok && keyRaw != nil {
			if keyStr, ok := keyRaw.(string); ok {
				attrValues["key"] = types.StringValue(keyStr)
			} else {
				attrValues["key"] = types.StringNull()
			}
		} else {
			attrValues["key"] = types.StringNull()
		}

		if opRaw, ok := itemMap["operator"]; ok && opRaw != nil {
			if opStr, ok := opRaw.(string); ok {
				attrValues["operator"] = types.StringValue(opStr)
			} else {
				attrValues["operator"] = types.StringNull()
			}
		} else {
			attrValues["operator"] = types.StringNull()
		}

		if valRaw, ok := itemMap["value"]; ok && valRaw != nil {
			if valStr, ok := valRaw.(string); ok {
				attrValues["value"] = types.StringValue(valStr)
			} else {
				attrValues["value"] = types.StringNull()
			}
		} else {
			attrValues["value"] = types.StringNull()
		}

		elemObj, objDiags := types.ObjectValue(labelConditionAttrTypes(), attrValues)
		if objDiags.HasError() {
			return types.ListNull(elemType), fmt.Errorf("creating label condition object: %s", objDiags.Errors()[0].Detail())
		}
		elements = append(elements, elemObj)
	}

	list, diags := types.ListValue(elemType, elements)
	if diags.HasError() {
		return types.ListNull(elemType), fmt.Errorf("creating label conditions list: %s", diags.Errors()[0].Detail())
	}
	return list, nil
}

// convertStringListFromClient converts a list of strings from API format to Terraform types.List.
// Returns null for nil or empty arrays to match Terraform's null semantics for unspecified fields.
func convertStringListFromClient(raw interface{}) (types.List, error) {
	if raw == nil {
		return types.ListNull(types.StringType), nil
	}

	rawList, ok := raw.([]interface{})
	if !ok {
		return types.ListNull(types.StringType), fmt.Errorf("expected list, got %T", raw)
	}

	// Return null for empty arrays to avoid plan/state mismatches
	// (API returns [] but Terraform config doesn't specify the field)
	if len(rawList) == 0 {
		return types.ListNull(types.StringType), nil
	}

	elements := make([]attr.Value, 0, len(rawList))
	for _, item := range rawList {
		if strVal, ok := item.(string); ok {
			elements = append(elements, types.StringValue(strVal))
		}
	}

	list, diags := types.ListValue(types.StringType, elements)
	if diags.HasError() {
		return types.ListNull(types.StringType), fmt.Errorf("creating string list: %s", diags.Errors()[0].Detail())
	}
	return list, nil
}

// RuleConditionsAttrTypes returns the attribute types for the conditions object.
// This is needed for creating null/unknown condition objects with the correct schema.
func RuleConditionsAttrTypes() map[string]attr.Type {
	// Define nested types for match conditions
	matchConditionAttrTypes := map[string]attr.Type{
		"match_on": types.ListType{ElemType: types.StringType},
		"operator": types.StringType,
	}

	// Define nested types for tag/label conditions
	labelConditionObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":      types.StringType,
			"operator": types.StringType,
			"value":    types.StringType,
		},
	}

	return map[string]attr.Type{
		"host_name":            types.ObjectType{AttrTypes: matchConditionAttrTypes},
		"host_tags":            types.ListType{ElemType: labelConditionObjectType},
		"host_labels":          types.ListType{ElemType: labelConditionObjectType},
		"host_label_groups":    types.ListType{ElemType: types.StringType},
		"service_description":  types.ObjectType{AttrTypes: matchConditionAttrTypes},
		"service_labels":       types.ListType{ElemType: labelConditionObjectType},
		"service_label_groups": types.ListType{ElemType: types.StringType},
	}
}
