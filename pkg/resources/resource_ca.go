package resources

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

const (
	kmsKeyID        = "kms_key_id"
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
			kmsKeyID: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The kms key with which we should encrypt the CA password.",
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

	password, err := generateRandomBytes(caPasswordBytes)
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

	d.Set("encrypted_ca", encoded.Bytes())
	return
}

func resourceCARead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCAUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCADelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate password")
	}
	return b, nil
}
