package provider

import (
	"crypto/x509"
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const ()

func KMSPublicKey() *schema.Resource {
	kmsPublicKey := newDataKMSPublicKey()

	return &schema.Resource{
		Read: kmsPublicKey.Read,

		Schema: map[string]*schema.Schema{
			schemaKmsKeyID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kms key we should get the public key",
			},

			// computed
			schemaPublicKey: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "This is the CA public key in openssh format",
			},
		},
	}
}

func newDataKMSPublicKey() *dataKMSPublicKey {
	return &dataKMSPublicKey{}
}

type dataKMSPublicKey struct{}

func (l *dataKMSPublicKey) Read(d *schema.ResourceData, meta interface{}) error {
	awsClient, ok := meta.(*aws.Client)
	if !ok {
		return errors.New("meta is not of type *aws.Client")
	}
	kmsKeyID := d.Get(schemaKmsKeyID).(string)

	svc := awsClient.KMS.Svc

	fmt.Printf("nil svc: %#v", svc == nil)

	output, err := svc.GetPublicKey(
		&kms.GetPublicKeyInput{KeyId: &kmsKeyID},
	)
	if err != nil {
		return errors.Wrap(err, "error getting kms public key")
	}
	pub, err := x509.ParsePKIXPublicKey(output.PublicKey)
	if err != nil {
		return errors.Wrap(err, "could not parse kms public key")
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return errors.Wrap(err, "could not ssh parse kms public key")
	}
	d.SetId(output.KeyId)
	d.Set(schemaPublicKey, string(ssh.MarshalAuthorizedKey(sshPub)))
	return nil
}
