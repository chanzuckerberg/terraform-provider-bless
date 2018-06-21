package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform/helper/schema"
)

// Client is an AWS client
type Client struct {
	KMS KMS
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
		KMS: NewKMS(kms.New(kmsSession)),
	}
	return client, nil
}
