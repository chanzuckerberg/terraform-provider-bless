package provider_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/provider"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type KMSMock struct {
	kmsiface.KMSAPI
	mock.Mock
}

func (k *KMSMock) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {
	args := k.Called(input)
	output := args.Get(0).(*kms.EncryptOutput)
	return output, args.Error(1)
}

func getTestProviders() (map[string]terraform.ResourceProvider, *KMSMock) {
	ca := provider.Provider()
	kmsMock := &KMSMock{}
	ca.ConfigureFunc = func(s *schema.ResourceData) (interface{}, error) {
		client := &aws.Client{
			KMS: aws.KMS{Svc: kmsMock},
		}
		return client, nil
	}
	providers := map[string]terraform.ResourceProvider{
		"bless": ca,
	}
	return providers, kmsMock
}

func TestProvider(t *testing.T) {
	assert := assert.New(t)
	p := provider.Provider()
	err := p.InternalValidate()
	assert.Nil(err)
}
