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
	svc kmsiface.KMSAPI
}

// NewKMS returns a new KMS client
func NewKMS(d *schema.ResourceData) (*KMS, error) {
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
		},
	))
	svc := kms.New(kmsSession)
	return &KMS{svc: svc}, nil
}
