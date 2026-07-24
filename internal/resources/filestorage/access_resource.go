package filestorage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &AccessResource{}
	_ resource.ResourceWithConfigure   = &AccessResource{}
	_ resource.ResourceWithImportState = &AccessResource{}
)

type accessClient interface {
	AssignAccess(ctx context.Context, fileStorageName, targetType, targetName string) (*client.FileStorageAccess, error)
	GetAccess(ctx context.Context, fileStorageName, targetType, targetName string) (*client.FileStorageAccess, error)
	RevokeAccess(ctx context.Context, fileStorageName, targetType, targetName string) error
}

// AccessResourceModel describes the access resource data model.
type AccessResourceModel struct {
	ID              types.String `tfsdk:"id"`
	FileStorageName types.String `tfsdk:"file_storage_name"`
	TargetType      types.String `tfsdk:"target_type"`
	TargetName      types.String `tfsdk:"target_name"`
}

// AccessResource defines the resource implementation.
type AccessResource struct {
	client accessClient
}

// NewAccessResource creates a new AccessResource.
func NewAccessResource() resource.Resource {
	return &AccessResource{}
}

// Configure adds the provider-configured client to the resource.
func (r *AccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if c.FileStorage == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected file storage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = c.FileStorage
}

// Metadata updates the resource type name.
func (r *AccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_access"
}

// Schema returns the schema for the resource.
func (r *AccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Grants a workload NFS access to a file storage in the ASC platform. " +
			"The backend resolves the workload's network CIDR and updates the NFS export configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this access entry ({file_storage_name}:{target_type}:{target_name}).",
				Computed:    true,
			},
			"file_storage_name": schema.StringAttribute{
				Description: "Name of the file storage to grant access to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Description: "Type of the workload to grant access to. Valid values: VirtualMachine, Container.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_name": schema.StringAttribute{
				Description: "Name of the workload to grant access to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create grants workload access to a file storage.
func (r *AccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AccessResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.AssignAccess(
		ctx,
		plan.FileStorageName.ValueString(),
		plan.TargetType.ValueString(),
		plan.TargetName.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error assigning file storage access",
			fmt.Sprintf("Could not assign access: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(accessStateID(plan.FileStorageName.ValueString(), plan.TargetType.ValueString(), plan.TargetName.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the access entry from the API and stores it in state.
func (r *AccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	access, err := r.client.GetAccess(
		ctx,
		state.FileStorageName.ValueString(),
		state.TargetType.ValueString(),
		state.TargetName.ValueString(),
	)
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading file storage access",
			fmt.Sprintf("Could not read access entry: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(accessStateID(access.FileStorageName, access.TargetType, access.TargetName))
	state.FileStorageName = types.StringValue(access.FileStorageName)
	state.TargetType = types.StringValue(access.TargetType)
	state.TargetName = types.StringValue(access.TargetName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported — all attributes require replacement.
func (r *AccessResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"File storage access updates are not supported. Changes require recreation.",
	)
}

// Delete revokes workload access from a file storage.
func (r *AccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AccessResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RevokeAccess(
		ctx,
		state.FileStorageName.ValueString(),
		state.TargetType.ValueString(),
		state.TargetName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error revoking file storage access",
			fmt.Sprintf("Could not revoke access: %s", err.Error()),
		)
		return
	}
}

// ImportState imports an access entry by composite key: "file-storage-name:target-type:target-name".
func (r *AccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format 'file-storage-name:target-type:target-name', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), accessStateID(parts[0], parts[1], parts[2]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("file_storage_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_name"), parts[2])...)
}

func accessStateID(fileStorageName, targetType, targetName string) string {
	return fileStorageName + ":" + targetType + ":" + targetName
}
