// Package tags provide helper function for translating tags between client and terraform
package tags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

// ToTerraform converts a client.Tag array into a terraform Map value
func ToTerraform(ctx context.Context, t []client.Tag, diags *diag.Diagnostics) types.Map {
	if len(t) == 0 {
		return types.MapNull(types.StringType)
	}
	tagsMap := make(map[string]string, len(t))
	for _, tag := range t {
		tagsMap[tag.Key] = tag.Value
	}
	tagsValue, d := types.MapValueFrom(ctx, types.StringType, tagsMap)
	diags.Append(d...)
	return tagsValue
}

// ToClient converts the Terraform tags model into client tags for the API request.
func ToClient(ctx context.Context, tags types.Map, diags *diag.Diagnostics) []client.Tag {
	var res []client.Tag
	if !tags.IsNull() && !tags.IsUnknown() {
		var tagsMap map[string]string
		diags.Append(tags.ElementsAs(ctx, &tagsMap, false)...)
		if diags.HasError() {
			return nil
		}
		for k, v := range tagsMap {
			res = append(res, client.Tag{Key: k, Value: v})
		}
	}
	return res
}
