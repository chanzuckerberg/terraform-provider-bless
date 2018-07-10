package aws

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
)

// KMS is a kms client
type KMS struct {
	Svc kmsiface.KMSAPI
}

// NewKMS returns a KMS client
func NewKMS(s *session.Session) KMS {
	return KMS{kms.New(s)}
}

// EncryptBytes encrypts the plaintext using the keyID key, result is base64 encoded
func (k *KMS) EncryptBytes(plaintext []byte, keyID string) (string, error) {
	input := &kms.EncryptInput{}
	input.SetKeyId(keyID).SetPlaintext(plaintext)
	response, err := k.Svc.Encrypt(input)

	return base64.StdEncoding.EncodeToString(response.CiphertextBlob),
		errors.Wrap(err, "Could not encrypt password")
}
