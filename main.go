// Package main implements the main entry point for the DSPC Terraform provider.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go tool tfplugindocs generate -provider-name dspc

var (
	version = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/NL-AMS-DSPC/dspc",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
