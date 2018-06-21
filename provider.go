package main

import (
	"github.com/chanzuckerberg/terraform-provider-bless-ca/pkg/aws"
	"github.com/chanzuckerberg/terraform-provider-bless-ca/pkg/resources"
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider is a provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				InputDefault: "us-east-1",
			},
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"allowed_account_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set:      schema.HashString,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bless_ca": resources.CA(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(s *schema.ResourceData) (interface{}, error) {
	return aws.NewClient(s)
}
