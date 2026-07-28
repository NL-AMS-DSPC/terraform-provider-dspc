package image

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockImageDataClient implements DataClient and returns a canned list of images.
type mockImageDataClient struct {
	response []client.ImageResponse
	err      error
}

func (m *mockImageDataClient) ListImages(_ context.Context) ([]client.ImageResponse, error) {
	return m.response, m.err
}

func TestRead(t *testing.T) {
	ctx := context.Background()

	var schemaResp datasource.SchemaResponse
	(&DataSource{}).Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	tests := map[string]struct {
		mockResponse []client.ImageResponse
		mockErr      error
		expectError  bool
	}{
		"populates state from client response": {
			mockResponse: []client.ImageResponse{
				{
					ID:                     "image-id",
					Name:                   "image-name",
					Family:                 "image-family",
					Distribution:           "image-distribution",
					Release:                "image-release",
					RequiresLicense:        true,
					LicenseInfo:            "image-license-info",
					SupportedArchitectures: []string{"arch1", "arch2"},
				},
			},
		},
		"empty result produces empty images list": {
			mockResponse: []client.ImageResponse{},
		},
		"client error becomes diagnostic error": {
			mockErr:     assert.AnError,
			expectError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			d := &DataSource{client: &mockImageDataClient{response: tt.mockResponse, err: tt.mockErr}}

			resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			d.Read(ctx, datasource.ReadRequest{}, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				return
			}
			require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

			var out DataSourceModel
			require.False(t, resp.State.Get(ctx, &out).HasError())
			require.Len(t, out.Images, len(tt.mockResponse))

			for i, want := range tt.mockResponse {
				img := out.Images[i]
				assert.Equal(t, want.ID, img.ID.ValueString())
				assert.Equal(t, want.Name, img.Name.ValueString())
				assert.Equal(t, want.Family, img.Family.ValueString())
				assert.Equal(t, want.Distribution, img.Distribution.ValueString())
				assert.Equal(t, want.Release, img.Release.ValueString())
				assert.Equal(t, want.RequiresLicense, img.RequiresLicense.ValueBool())
				assert.Equal(t, want.LicenseInfo, img.LicenseInfo.ValueString())

				var archs []string
				require.False(t, img.SupportedArchitectures.ElementsAs(ctx, &archs, false).HasError())
				assert.Equal(t, want.SupportedArchitectures, archs)
			}
		})
	}
}
