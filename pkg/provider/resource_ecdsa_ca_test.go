package provider_test

import (
	"crypto/rand"
	"encoding/base64"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	r "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateECDSA(t *testing.T) {
	a := assert.New(t)
	providers, kmsMock := getTestProviders()

	ciphertext := make([]byte, 10)
	rand.Read(ciphertext)
	output := &kms.EncryptOutput{
		CiphertextBlob: ciphertext,
	}
	kmsMock.On("Encrypt", mock.Anything).Return(output, nil)

	r.Test(t, r.TestCase{
		Providers: providers,
		Steps: []r.TestStep{
			r.TestStep{
				Config: `
				provider "bless" {
					region = "us-east-1"
				}

				resource "bless_ecdsa_ca" "bless" {
					kms_key_id = "testo"
				}

				output "ecdsa_private_key" {
					value = "${bless_ecdsa_ca.bless.encrypted_ca}"
				}
				output "ecdsa_public_key" {
					value = "${bless_ecdsa_ca.bless.public_key}"
				}
				output "ecdsa_password" {
					value = "${bless_ecdsa_ca.bless.encrypted_password}"
				}
			`,
				Check: func(s *terraform.State) error {
					privateUntyped := s.RootModule().Outputs["ecdsa_private_key"].Value
					private, ok := privateUntyped.(string)
					a.True(ok)
					bytesPrivate, err := base64.StdEncoding.DecodeString(private)
					a.Nil(err)
					a.Regexp(
						regexp.MustCompile("^-----BEGIN EC PRIVATE KEY-----"),
						string(bytesPrivate))
					a.Regexp(
						regexp.MustCompile(`AES-256-CBC`),
						string(bytesPrivate))
					publicSSHUntyped := s.RootModule().Outputs["ecdsa_public_key"].Value
					publicSSH, ok := publicSSHUntyped.(string)
					a.True(ok)
					a.Regexp(
						regexp.MustCompile("^ecdsa-sha2-nistp521 "),
						string(publicSSH))
					return nil
				},
			},
		},
	})
}
