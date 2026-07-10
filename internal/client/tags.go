package client

// Tag represents a key-value pair used for tagging resources.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PatchTagDTO represents a single tag mutation in a patch request
type PatchTagDTO struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

// PatchTagsRequest is the request body for partial tag updates
type PatchTagsRequest struct {
	Tags []PatchTagDTO `json:"tags,omitempty"`
}

// TagDTO is the request-side representation of a tag
type TagDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
