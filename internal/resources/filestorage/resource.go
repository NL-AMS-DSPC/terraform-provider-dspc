// Package filestorage implements Terraform resources and data sources for the ASC file storage service.
package filestorage

import (
	"context"
	"errors"
	"fmt"

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
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type fsClient interface {
	CreateFileStorage(ctx context.Context, name, size string) (*client.FileStorage, error)
	GetFileStorage(ctx context.Context, name string) (*client.FileStorage, error)
	DeleteFileStorage(ctx context.Context, name string) error
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Size         types.String `tfsdk:"size"`
	Status       types.String `tfsdk:"status"`
	NFSMountPath types.String `tfsdk:"nfs_mount_path"`
}

// Resource defines the resource implementation.
type Resource struct {
	client fsClient
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Configure adds the provider-configured client to the resource.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage"
}

// Schema returns the schema for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a file storage volume in the ASC platform. " +
			"File storages are CephFS-backed NFS shares that can be mounted by workloads.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the file storage (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the file storage. Must be unique within the platform.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.StringAttribute{
				Description: "Size of the file storage (e.g. 100Gi).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the file storage.",
				Computed:    true,
			},
			"nfs_mount_path": schema.StringAttribute{
				Description: "NFS path used to mount the file storage.",
				Computed:    true,
			},
		},
	}
}

// Create creates a new file storage.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fs, err := r.client.CreateFileStorage(ctx, plan.Name.ValueString(), plan.Size.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating file storage",
			fmt.Sprintf("Could not create file storage: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fs.Name)
	plan.Name = types.StringValue(fs.Name)
	plan.Size = types.StringValue(fs.Size)
	plan.Status = types.StringValue(fs.Status)
	plan.NFSMountPath = types.StringValue(fs.NFSMountPath)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the file storage from the API and stores it in state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fs, err := r.client.GetFileStorage(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading file storage",
			fmt.Sprintf("Could not read file storage: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(fs.Name)
	state.Name = types.StringValue(fs.Name)
	state.Size = types.StringValue(fs.Size)
	state.Status = types.StringValue(fs.Status)
	state.NFSMountPath = types.StringValue(fs.NFSMountPath)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported — all attributes require replacement.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"File storage updates are not supported. Changes require recreation.",
	)
}

// Delete deletes the file storage.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFileStorage(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting file storage",
			fmt.Sprintf("Could not delete file storage: %s", err.Error()),
		)
		return
	}
}

// ImportState imports a file storage by name.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
