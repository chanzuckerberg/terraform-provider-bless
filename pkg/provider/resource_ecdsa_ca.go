package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/util"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

// ECDSACA is an ecdsa CA resource
func ECDSACA() *schema.Resource {
	ca := newResourceECDSACA()
	return &schema.Resource{
		Create: ca.Create,
		Read:   ca.Read,
		Delete: ca.Delete,

		Schema: map[string]*schema.Schema{
			schemaKmsKeyID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kms key with which we should encrypt the CA password.",
				ForceNew:    true,
			},

			// computed
			schemaEncryptedPrivateKey: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "This is the base64 encoded CA encrypted private key.",
			},
			schemaPublicKey: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "This is the plaintext CA public key in openssh format.",
			},
			schemaEncryptedPassword: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "This is the kms encrypted password.",
			},
		},
	}
}

// resourceCA is a namespace
type resourceECDSACA struct{}

func newResourceECDSACA() *resourceECDSACA {
	return &resourceECDSACA{}
}

// Create creates a CA
func (ca *resourceECDSACA) Create(d *schema.ResourceData, meta interface{}) error {
	awsClient, ok := meta.(*aws.Client)
	if !ok {
		return errors.New("meta is not of type *aws.Client")
	}

	kmsKeyID := d.Get(schemaKmsKeyID).(string)
	keyPair, err := ca.createKeypair()
	if err != nil {
		return err
	}

	encryptedPassword, err := awsClient.KMS.EncryptBytes(keyPair.Password, kmsKeyID)
	if err != nil {
		return err
	}

	d.Set(schemaEncryptedPrivateKey, keyPair.B64EncryptedPrivateKey)
	d.Set(schemaPublicKey, keyPair.PublicKey)
	d.Set(schemaEncryptedPassword, encryptedPassword)
	d.SetId(util.HashForState(keyPair.PublicKey))
	return nil
}

// Read reads the ca
func (ca *resourceECDSACA) Read(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// Delete deletes the ca
func (ca *resourceECDSACA) Delete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

// ------------ helpers ------------------
func (ca *resourceECDSACA) createKeypair() (*util.CA, error) {
	// generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "Private key generation failed")
	}
	return util.NewCA(privateKey, privateKey.Public(), caPasswordBytes)
}
