package provider

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/util"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
)

const (
	schemaKmsKeyID            = "kms_key_id"
	schemaEncryptedPrivateKey = "encrypted_ca"
	schemaPublicKey           = "public_key"
	schemaEncryptedPassword   = "encrypted_password"

	keySize         = 4096
	caPasswordBytes = 64
)

// CA is a bless CA resource
func CA() *schema.Resource {
	ca := newResourceCA()
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
type resourceCA struct{}

func newResourceCA() *resourceCA {
	return &resourceCA{}
}

// Create creates a CA
func (ca *resourceCA) Create(d *schema.ResourceData, meta interface{}) error {
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

	d.Set(schemaEncryptedPrivateKey, keyPair.B64EncryptedPrivateKey) // nolint
	d.Set(schemaPublicKey, keyPair.PublicKey) // nolint
	d.Set(schemaEncryptedPassword, encryptedPassword) // nolint
	d.SetId(util.HashForState(keyPair.PublicKey))
	return nil
}

// Read reads the ca
func (ca *resourceCA) Read(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// Delete deletes the ca
func (ca *resourceCA) Delete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

// ------------ helpers ------------------
func (ca *resourceCA) createKeypair() (*util.CA, error) {
	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, errors.Wrap(err, "Private key generation failed")
	}
	return util.NewCA(privateKey, privateKey.Public(), caPasswordBytes)
}
