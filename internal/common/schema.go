package common

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// ComputedIDAttribute returns a standard computed ID attribute
func ComputedIDAttribute(description string) schema.StringAttribute {
	if description == "" {
		description = "The unique identifier for the resource."
	}
	return schema.StringAttribute{
		MarkdownDescription: description,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// RequiredIDAttribute returns a standard required ID attribute that requires replace on change
func RequiredIDAttribute(description string) schema.StringAttribute {
	if description == "" {
		description = "Unique identifier for the resource."
	}
	return schema.StringAttribute{
		MarkdownDescription: description,
		Required:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// StrictResourceLockingAttribute returns the standard strict_resource_locking attribute
func StrictResourceLockingAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Override provider-level strict_resource_locking for this resource. " +
			"Controls ETag validation for resource operations. " +
			"If not set, uses provider configuration.",
		Optional: true,
	}
}

// ActivateAttribute returns the standard activate attribute for resource-level override
func ActivateAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Override provider-level activation mode for this resource. " +
			"Options: 'auto', 'manual'. If not set, uses provider configuration.",
		Optional: true,
	}
}

// ForceForeignChangesAttribute returns the standard force_foreign_changes attribute
func ForceForeignChangesAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Override provider-level force_foreign_changes for this resource. " +
			"If not set, uses provider configuration.",
		Optional: true,
	}
}
