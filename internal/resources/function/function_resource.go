package function

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &FunctionResource{}
	_ resource.ResourceWithConfigure   = &FunctionResource{}
	_ resource.ResourceWithImportState = &FunctionResource{}
)

// FunctionResourceClient defines the interface for managing function resources.
// It provides methods to create, delete, retrieve, and list functions.
type FunctionResourceClient interface {
	CreateFunction(ctx context.Context, name, image string) (*client.Function, error)
	DeleteFunction(ctx context.Context, name string) error
	GetFunction(ctx context.Context, name string) (*client.Function, error)
	ListFunctions(ctx context.Context) ([]*client.Function, error)
}

// FunctionResource defines the resource implementation.
type FunctionResource struct {
	client FunctionResourceClient
}

// FunctionResourceModel describes the resource data model.
type FunctionResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Image  types.String `tfsdk:"image"`
	Status types.String `tfsdk:"status"`
}

// NewFunctionResource creates a new FunctionResource.
func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *FunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *FunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a function in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the function.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the function. Must be unique within the platform.",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "The container image for the function (e.g., 'gcr.io/knative-samples/helloworld-go').",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the function (e.g., \"pending\", \"ready\").",
				Computed:    true,
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *FunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Functions == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected functions service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Functions
}

// Create creates a new function in the DSPC platform.
func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FunctionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the function using the provider-level namespace
	function, err := r.client.CreateFunction(ctx, plan.Name.ValueString(), plan.Image.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating function",
			fmt.Sprintf("Could not create function: %s", err.Error()),
		)
		return
	}

	// Set the computed values
	plan.ID = types.StringValue(function.Name)
	plan.Status = types.StringValue(function.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FunctionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	functionName := state.Name.ValueString()

	// Get the function using the provider-level namespace
	function, err := r.client.GetFunction(ctx, functionName)

	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			// If function not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting function",
			fmt.Sprintf("Could not get function: %s", err.Error()),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(function.Name)
	state.Status = types.StringValue(function.Status)
	// Note: function.Name and function.Image stay as they are

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the function in the DSPC platform.
func (r *FunctionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since the API only supports function name and image and doesn't have update operations,
	// we treat any changes as requiring recreation (ForceNew)
	resp.Diagnostics.AddError(
		"Update not supported",
		"Function updates are not supported by the DSPC API. Changes require function recreation.",
	)
}

// Delete deletes the function in the DSPC platform.
func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FunctionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	functionName := state.Name.ValueString()

	// Delete the function using the provider-level namespace
	err := r.client.DeleteFunction(ctx, functionName)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting function",
			fmt.Sprintf("Could not delete function: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the function in the DSPC platform.
func (r *FunctionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
