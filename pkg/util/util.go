package util

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// HashForState hashes the state
func HashForState(value string) string {
	if value == "" {
		return ""
	}
	hash := sha1.Sum([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(hash[:])
}

// HashFileForState hashes a file's contents
func HashFileForState(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "Could not open file %s", path)
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", errors.Wrap(err, "Could not add file contents to hash")
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(n uint) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate password")
	}
	return b, nil
}

// CA has information around a CA
type CA struct {
	PublicKey              string
	B64EncryptedPrivateKey string
	Password               []byte
}

// NewCA generates the CA components from a private key
func NewCA(privateKey crypto.PrivateKey, publicKey interface{}, passwordBytes uint) (*CA, error) {
	block := &pem.Block{}

	switch typed := privateKey.(type) {
	case *ecdsa.PrivateKey:
		block.Type = "EC PRIVATE KEY"
		bytes, err := x509.MarshalECPrivateKey(typed)
		if err != nil {
			return nil, errors.Wrap(err, "Could not x509 Marshal private key")
		}
		block.Bytes = bytes
	case *rsa.PrivateKey:
		block.Type = "RSA PRIVATE KEY"
		block.Bytes = x509.MarshalPKCS1PrivateKey(typed)
	default:
		return nil, errors.New("Unrecognized private key type")
	}

	password, err := GenerateRandomBytes(passwordBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate private key password")
	}

	// encrypt the private key
	block, err = x509.EncryptPEMBlock(
		rand.Reader,
		block.Type,
		block.Bytes,
		password,
		x509.PEMCipherAES256)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to encrypt CA")
	}
	var encryptedPEMBytes bytes.Buffer
	err = pem.Encode(&encryptedPEMBytes, block)
	if err != nil {
		return nil, errors.Wrap(err, "Could not PEM encode encrypted CA")
	}

	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate openssh public key")
	}
	return &CA{
		PublicKey:              string(ssh.MarshalAuthorizedKey(sshPublicKey)),
		B64EncryptedPrivateKey: base64.StdEncoding.EncodeToString(encryptedPEMBytes.Bytes()),
		Password:               password,
	}, nil
}
