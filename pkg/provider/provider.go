package provider

import (
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
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
		},
		ResourcesMap: map[string]*schema.Resource{
			"bless_ca": CA(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(s *schema.ResourceData) (interface{}, error) {
	return aws.NewClient(s)
}
