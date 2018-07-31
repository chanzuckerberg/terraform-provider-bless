package provider_test

import (
	"testing"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/provider"
	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
)

func TestLambdaCreate(t *testing.T) {
	a := assert.New(t)

	providers, _ := getTestProviders()

	r.Test(t, r.TestCase{
		Providers: providers,
		Steps: []r.TestStep{
			r.TestStep{
				Config: `
				provider "bless" {
					region = "us-east-1"
				}

				data "bless_lambda" "zip" {
					encrypted_ca = "aaaa"
					encrypted_password = "bbbb"
					service_name = "test"
					kmsauth_key_id = "keyID"

				}

				output "output_path" {
					value = "${data.bless_lambda.zip.output_path}"
				}

				output "output_base64sha256" {
					value = "${data.bless_lambda.zip.output_base64sha256}"
				}
				`,
				Check: func(s *terraform.State) error {
					outputPath := s.RootModule().Outputs[provider.SchemaOutputPath].Value
					base64sha256 := s.RootModule().Outputs[provider.SchemaOutputBase64Sha256].Value
					a.NotEmpty(outputPath)
					a.NotEmpty(base64sha256)
					return nil
				},
			},
		},
	})
}
