package provider

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/chanzuckerberg/terraform-provider-bless/pkg/aws"
	"github.com/chanzuckerberg/terraform-provider-bless/pkg/util"
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

func newResourceCA() resourceCA {
	return resourceCA{}
}

type keyPair struct {
	publicKey              string
	b64EncryptedPrivateKey string
	password               []byte
}

// Create creates a CA
func (ca resourceCA) Create(d *schema.ResourceData, meta interface{}) error {
	awsClient, ok := meta.(*aws.Client)
	if !ok {
		return errors.New("meta is not of type *aws.Client")
	}

	kmsKeyID := d.Get(schemaKmsKeyID).(string)
	keyPair, err := ca.createKeypair()
	if err != nil {
		return err
	}
	encryptedPassword, err := awsClient.KMS.EncryptBytes(keyPair.password, kmsKeyID)
	if err != nil {
		return err
	}

	d.Set(schemaEncryptedPrivateKey, keyPair.b64EncryptedPrivateKey)
	d.Set(schemaPublicKey, keyPair.publicKey)
	d.Set(schemaEncryptedPassword, encryptedPassword)
	d.SetId(util.HashForState(keyPair.publicKey))
	return nil
}

// Read reads the ca
func (ca resourceCA) Read(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// Delete deletes the ca
func (ca resourceCA) Delete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

// ------------ helpers ------------------
func (ca resourceCA) createKeypair() (*keyPair, error) {
	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, errors.Wrap(err, "Private key generation failed")
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// generate password
	password, err := util.GenerateRandomBytes(caPasswordBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate enough random bytes")
	}

	// encrypt the private key
	block, err = x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, password, x509.PEMCipherAES256)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encrypt CA")
	}
	var encoded bytes.Buffer
	err = pem.Encode(&encoded, block)
	if err != nil {
		return nil, errors.Wrap(err, "Could not PEM encode encrypted CA")
	}

	// public key in openssh format
	sshPubKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate openssh public key")
	}

	return &keyPair{
		publicKey:              string(ssh.MarshalAuthorizedKey(sshPubKey)),
		b64EncryptedPrivateKey: base64.StdEncoding.EncodeToString(encoded.Bytes()),
		password:               password,
	}, nil
}
