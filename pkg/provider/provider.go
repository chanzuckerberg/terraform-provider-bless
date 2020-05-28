package provider

import (
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"role_arn"},
			},
			"role_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"profile"},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"bless_ca":       CA(),
			"bless_ecdsa_ca": ECDSACA(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"bless_lambda":         Lambda(),
			"bless_kms_public_key": KMSPublicKey(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(s *schema.ResourceData) (interface{}, error) {
	return aws.NewClient(s)
}
