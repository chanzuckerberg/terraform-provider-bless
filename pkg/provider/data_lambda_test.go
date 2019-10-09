package provider_test

import (
	"archive/zip"
	"bufio"
	"strings"
	"testing"

	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					encrypted_password = "bb+bb"
					service_name = "test"
					kmsauth_key_id = "keyID"
					output_path = "/tmp/test.zip"
				}

				data "bless_lambda" "zip2" {
					encrypted_ca = "aaaa"
					encrypted_password = "bb+bb"
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
					output1 := s.RootModule().Outputs["output"].Value
					output2 := s.RootModule().Outputs["output_2"].Value
					a.NotEmpty(output1)
					a.NotEmpty(output2)
					// Check hashes are equal

					validateBlessConfig(t, "/tmp/test.zip")
					a.Equal(output1, output2)
					return nil
				},
				Destroy: true,
			},
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
					output_path = "/tmp/test3.zip"
				}

				data "bless_lambda" "zip2" {
					encrypted_ca = "aaaa"
					encrypted_password = "bbbb"
					service_name = "test"
					kmsauth_key_id = "keyID"
					output_path = "/tmp/test4.zip"
					kmsauth_validate_remote_user = "false" # setting different field here
				}


				output "output" {
					value = "${data.bless_lambda.zip.output_base64sha256}"
				}
				output "output_2" {
					value = "${data.bless_lambda.zip2.output_base64sha256}"
				}
				`,
				Check: func(s *terraform.State) error {
					output1 := s.RootModule().Outputs["output"].Value
					output2 := s.RootModule().Outputs["output_2"].Value
					a.NotEmpty(output1)
					a.NotEmpty(output2)
					// Check hashes are equal
					a.NotEqual(output1, output2)
					return nil
				},
				Destroy: true,
			},
		},
	})
}

func validateBlessConfig(t *testing.T, zipPath string) {
	a := assert.New(t)

	r, err := zip.OpenReader(zipPath)
	a.Nil(err)
	defer r.Close()

	configFound := false
	privateKeyFound := false

	for _, f := range r.File {
		if f.Name != "bless_deploy.cfg" {
			continue
		}
		configFound = true
		fc, err := f.Open()
		a.Nil(err)
		defer fc.Close()
		scanner := bufio.NewScanner(fc)

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "bb+bb") {
				privateKeyFound = true
				break
			}
		}
		a.Nil(scanner.Err())

		break // Found the config file
	}

	a.True(configFound)
	a.True(privateKeyFound)
}
