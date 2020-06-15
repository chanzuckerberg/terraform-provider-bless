package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/provider"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/version"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func main() {
	ver := flag.Bool("version", false, "spit out version for resources here")
	flag.Parse()

	if *ver {
		verString, err := version.VersionString()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(verString)
		return
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return provider.Provider()
		},
	})

}
