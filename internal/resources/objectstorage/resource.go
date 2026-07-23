package objectstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type quotaModel struct {
	MaxSize types.String `tfsdk:"max_size"`
}

type tagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type resourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	TenantID      types.String `tfsdk:"tenant_id"`
	ReclaimPolicy types.String `tfsdk:"reclaim_policy"`
	Endpoint      types.String `tfsdk:"endpoint"`
	Region        types.String `tfsdk:"region"`
	Quota         quotaModel   `tfsdk:"quota"`
	Tags          tagModel     `tfsdk:"tags"`
}

type objectStorageClient interface {
	CreateBucket(ctx context.Context, req client.CreateObjectStorageRequest) (*client.ObjectStorage, error)
	GetBucket(ctx context.Context, id string) (*client.ObjectStorage, error)
	UpdateBucket(ctx context.Context, id string, req client.UpdateBucketRequest) (*client.ObjectStorage, error)
	DeleteBucket(ctx context.Context, id string) error
}

var objectStorageResourceSchema = schema.Schema{
	Description: "Retrieves an existing object storage from the ASC platform.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier of the object storage.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name of the object storage.",
			Required:    true,
		},
		"tenant_id": schema.StringAttribute{
			Description: "Identifier of the tenant that owns the object storage.",
			Computed:    true,
		},
		"reclaim_policy": schema.StringAttribute{
			Description: "Reclaim policy of the object storage.",
			Required:    true,
		},
		"endpoint": schema.StringAttribute{
			Description: "Endpoint of the object storage.",
			Computed:    true,
		},
		"region": schema.StringAttribute{
			Description: "Region of the object storage.",
			Required:    true,
		},
		"tags": schema.ObjectAttribute{
			Optional: true,
			AttributeTypes: map[string]attr.Type{
				"key":   types.StringType,
				"value": types.StringType,
			},
		},
	},
	Blocks: map[string]schema.Block{
		"quota": schema.SingleNestedBlock{
			Description: "the quota configuration for the object storage",
			Attributes: map[string]schema.Attribute{
				"max_size": schema.StringAttribute{
					Computed:    true,
					Description: "the max size of the object storage",
				},
			},
		},
	},
}

// Resource defines the resource implementation.
type Resource struct {
	client objectStorageClient
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

	if c.ObjectStorage == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected object storage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = c.ObjectStorage
}

// Metadata updates the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage"
}

// Schema returns the schema for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = objectStorageResourceSchema
}

// Create creates a new object storage.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectStorage, err := r.client.CreateBucket(ctx, client.CreateObjectStorageRequest{
		Name:          plan.Name.ValueString(),
		Quota:         &client.StorageQuota{MaxSize: plan.Quota.MaxSize.ValueString()},
		ReclaimPolicy: plan.ReclaimPolicy.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating object storage",
			fmt.Sprintf("Could not create object storage: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(objectStorage.ID)
	plan.Name = types.StringValue(objectStorage.Name)
	plan.TenantID = types.StringValue(objectStorage.TenantID)
	plan.Endpoint = types.StringValue(objectStorage.Endpoint)
	plan.Region = types.StringValue(objectStorage.Region)
	plan.ReclaimPolicy = types.StringValue(objectStorage.ReclaimPolicy)
	plan.Quota.MaxSize = types.StringValue(objectStorage.Quota.MaxSize)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the object storage from the API and stores it in state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.GetBucket(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading object storage",
			fmt.Sprintf("Could not read object storage: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(bucket.ID)
	state.Name = types.StringValue(bucket.Name)
	state.TenantID = types.StringValue(bucket.TenantID)
	state.Endpoint = types.StringValue(bucket.Endpoint)
	state.Region = types.StringValue(bucket.Region)
	state.ReclaimPolicy = types.StringValue(bucket.ReclaimPolicy)
	state.Quota.MaxSize = types.StringValue(bucket.Quota.MaxSize)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates a given object storage
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Quota.MaxSize.Equal(state.Quota.MaxSize) {
		_, err := r.client.UpdateBucket(ctx, state.ID.ValueString(), client.UpdateBucketRequest{
			Quota: client.StorageQuota{MaxSize: plan.Quota.MaxSize.ValueString()},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating object storage",
				fmt.Sprintf("Could not update object storage: %s", err.Error()),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the object storage.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBucket(ctx, state.ID.ValueString()); err != nil {
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
