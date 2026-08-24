// Package wrappers provides typed wrappers around CheckMK rule resources.
// This file re-exports condition utilities from the parent rules package
// to avoid code duplication and ensure consistency.
package wrappers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/withakedo/terraform_checkmk_provider/internal/rules"
)

// Re-export condition utilities from the rules package.
// These are used by rule wrappers to validate and convert conditions.

// ValidateRulesetConditions validates that conditions are appropriate for the ruleset.
// Returns errors for invalid combinations that would be silently dropped by the API.
var ValidateRulesetConditions = rules.ValidateRulesetConditions

// RuleConditionsSchema returns the Terraform schema for rule conditions.
func RuleConditionsSchema() schema.SingleNestedAttribute {
	return rules.RuleConditionsSchema()
}

// RuleConditionsAttrTypes returns the attribute types for the conditions object.
func RuleConditionsAttrTypes() map[string]attr.Type {
	return rules.RuleConditionsAttrTypes()
}

// ConvertConditionsToClient converts Terraform conditions to client format.
func ConvertConditionsToClient(ctx context.Context, conditions types.Object) (map[string]interface{}, error) {
	return rules.ConvertConditionsToClient(ctx, conditions)
}

// ConvertConditionsFromClient converts client conditions to Terraform format.
func ConvertConditionsFromClient(ctx context.Context, conditions map[string]interface{}) (types.Object, error) {
	return rules.ConvertConditionsFromClient(ctx, conditions)
}
