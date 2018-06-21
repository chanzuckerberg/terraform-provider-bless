package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/hashicorp/terraform/helper/schema"
)

// KMS is a kms client
type KMS struct {
	kmsiface.KMSAPI
}

// Client is an AWS client
type Client struct {
	KMS
}

// NewClient returns a new aws client
func NewClient(d *schema.ResourceData) (*Client, error) {
	regionOverride := d.Get("region").(string)
	var region *string
	if regionOverride != "" {
		region = aws.String(regionOverride)
	}
	kmsSession := session.Must(session.NewSessionWithOptions(
		session.Options{
			Config: aws.Config{
				Region: region,
			},
			SharedConfigState: session.SharedConfigEnable,
			Profile:           d.Get("profile").(string),
		},
	))
	client := &Client{
		KMS: KMS{kms.New(kmsSession)},
	}
	return client, nil
}
