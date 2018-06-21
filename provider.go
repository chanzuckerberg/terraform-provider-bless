package main

import (
	"github.com/chanzuckerberg/terraform-provider-bless-ca/pkg/resources"
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider is a provider
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"ca": resources.CA(),
		},
	}
}
