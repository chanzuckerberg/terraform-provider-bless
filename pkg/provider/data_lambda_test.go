package provider_test

import (
	"testing"

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
					output_path = "/tmp/test.zip"
				}

				data "bless_lambda" "zip2" {
					encrypted_ca = "aaaa"
					encrypted_password = "bbbb"
					service_name = "test"
					kmsauth_key_id = "keyID"
					output_path = "/tmp/test2.zip"
				}


				output "output" {
					value = "${data.bless_lambda.zip.output_base64sha256}"
				}
				output "output_2" {
					value = "${data.bless_lambda.zip2.output_base64sha256}"
				}

				`,
				Check: func(s *terraform.State) error {
					output1:= s.RootModule().Outputs["output"].Value
					output2:= s.RootModule().Outputs["output_2"].Value
					a.NotEmpty(output1)
					a.NotEmpty(output2)
					// Check hashes are equal
					a.Equal(output1, output2)
					return nil
				},
				Destroy: true,
			},
		},
	})
}
