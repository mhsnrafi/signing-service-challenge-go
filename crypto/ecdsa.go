package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
)

// ECCKeyPair is a DTO that holds ECC private and public keys.
type ECCKeyPair struct {
	Public  *ecdsa.PublicKey
	Private *ecdsa.PrivateKey
}

// ECCMarshaler can encode and decode an ECC key pair.
type ECCMarshaler struct{}

// NewECCMarshaler creates a new ECCMarshaler.
func NewECCMarshaler() ECCMarshaler {
	return ECCMarshaler{}
}

// Marshal takes an ECCKeyPair and encodes it to be written on disk.
// It returns the public and the private key as a byte slice.
func (m ECCMarshaler) Marshal(keyPair ECCKeyPair) ([]byte, []byte, error) {
	privateKeyBytes, err := x509.MarshalECPrivateKey(keyPair.Private)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(keyPair.Public)
	if err != nil {
		return nil, nil, err
	}

	encodedPrivate := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE_KEY",
		Bytes: privateKeyBytes,
	})

	encodedPublic := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC_KEY",
		Bytes: publicKeyBytes,
	})

	return encodedPublic, encodedPrivate, nil
}

// Unmarshal assembles an ECCKeyPair from an encoded private key.
func (m ECCMarshaler) Unmarshal(privateKeyBytes []byte) (*ECCKeyPair, error) {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &ECCKeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}, nil
}

type ECDSASignature struct {
	R, S *big.Int
}

type ECCSigner struct {
	PrivateKey *ecdsa.PrivateKey
}

func (s *ECCSigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	if s.PrivateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	hash := sha256.Sum256(dataToBeSigned)

	r, ss, err := ecdsa.Sign(rand.Reader, s.PrivateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	signature, err := asn1.Marshal(ECDSASignature{R: r, S: ss})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signature: %w", err)
	}

	return signature, nil
}
