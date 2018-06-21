package aws

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
)

// KMS is a kms client
type KMS struct {
	svc kmsiface.KMSAPI
}

// EncryptBytes encrypts the plaintext using the keyID key, result is base64 encoded
func (k *KMS) EncryptBytes(plaintext []byte, keyID string) (string, error) {
	input := &kms.EncryptInput{}
	input.SetKeyId(keyID).SetPlaintext(plaintext)
	response, err := k.svc.Encrypt(input)
	return string(response.CiphertextBlob), errors.Wrap(err, "Could not encrypt password")
}
