package resources

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/chanzuckerberg/terraform-provider-bless-ca/pkg/util"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const (
	schemaKmsKeyID            = "kms_key_id"
	schemaEncryptedPrivateKey = "encrypted_ca"
	schemaPublicKey           = "public_key"
	schemaEncryptedPassword   = "encrypted_password"

	keySize         = 4096
	caPasswordBytes = 32
)

// CA is a bless CA resource
func CA() *schema.Resource {
	return &schema.Resource{
		Create: resourceCACreate,
		Read:   resourceCARead,
		Update: resourceCAUpdate,
		Delete: resourceCADelete,

		Schema: map[string]*schema.Schema{
			schemaKmsKeyID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kms key with which we should encrypt the CA password.",
			},

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

func resourceCACreate(d *schema.ResourceData, m interface{}) error {
	// keyID := d.Get(kmsKeyID).(string)
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return errors.Wrap(err, "Private key generation failed")
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	password, err := util.GenerateRandomBytes(caPasswordBytes)
	if err != nil {
		return err
	}

	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, password, x509.PEMCipherAES256)
	if err != nil {
		return errors.Wrap(err, "Failed to encrypt CA")
	}

	var encoded bytes.Buffer
	err = pem.Encode(&encoded, block)
	if err != nil {
		return errors.Wrap(err, "Could not PEM encode encrypted CA")
	}
	d.Set(
		schemaEncryptedPrivateKey,
		base64.StdEncoding.EncodeToString(encoded.Bytes()))

	sshPubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return errors.Wrap(err, "Could not generate openssh public key")
	}
	sshPubKeyBytes := ssh.MarshalAuthorizedKey(sshPubKey)
	d.Set(schemaPublicKey, string(sshPubKeyBytes))
	return nil
}

func resourceCARead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCAUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCADelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
