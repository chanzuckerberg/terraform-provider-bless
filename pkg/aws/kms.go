package aws

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
)

// KMS is a kms client
type KMS struct {
	kmsiface.KMSAPI
}

// NewKMS returns a KMS client
func NewKMS(k kmsiface.KMSAPI) KMS {
	return KMS{k}
}

// EncryptBytes encrypts the plaintext using the keyID key, result is base64 encoded
func (k *KMS) EncryptBytes(plaintext []byte, keyID string) (string, error) {
	input := &kms.EncryptInput{}
	input.SetKeyId(keyID).SetPlaintext(plaintext)
	response, err := k.Encrypt(input)

	return base64.StdEncoding.EncodeToString(response.CiphertextBlob),
		errors.Wrap(err, "Could not encrypt password")
}
