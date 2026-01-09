package resources

import (
	"context"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestBlockStorageAttachmentResource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewBlockStorageAttachmentResource()

	schemaReq := tfresource.SchemaRequest{}
	schemaResp := &tfresource.SchemaResponse{}

	r.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Schema.Attributes == nil {
		t.Fatal("Expected schema attributes to be defined")
	}

	if _, ok := schemaResp.Schema.Attributes["id"]; !ok {
		t.Error("Expected 'id' attribute in schema")
	}

	if _, ok := schemaResp.Schema.Attributes["vm_name"]; !ok {
		t.Error("Expected 'vm_name' attribute in schema")
	}

	if _, ok := schemaResp.Schema.Attributes["block_storage_name"]; !ok {
		t.Error("Expected 'block_storage_name' attribute in schema")
	}
}
