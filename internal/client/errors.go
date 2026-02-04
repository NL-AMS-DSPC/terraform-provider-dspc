package client

import "errors"

var (
	// ErrResourceNotFound error is used for when a resource is not found
	ErrResourceNotFound = errors.New("resource not found")
)
