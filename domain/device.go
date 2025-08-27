package domain

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/crypto"
	"github.com/google/uuid"
)

type SignatureAlgorithm string

const (
	RSA SignatureAlgorithm = "RSA"
	ECC SignatureAlgorithm = "ECC"
)

type SignatureDevice struct {
	ID               string             `json:"id"`
	Label            string             `json:"label"`
	Algorithm        SignatureAlgorithm `json:"algorithm"`
	SignatureCounter int                `json:"signature_counter"`
	LastSignature    string             `json:"last_signature"`
	PublicKey        []byte             `json:"public_key"`
	PrivateKey       []byte             `json:"-"`
	mu               sync.Mutex
}

func NewSignatureDevice(id string, algorithm SignatureAlgorithm, label string) (*SignatureDevice, error) {
	if id == "" {
		return nil, errors.New("device ID cannot be empty")
	}

	if algorithm != RSA && algorithm != ECC {
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	var publicKey, privateKey []byte

	switch algorithm {
	case RSA:
		generator := &crypto.RSAGenerator{}
		keyPair, err := generator.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}

		marshaler := crypto.NewRSAMarshaler()
		publicKey, privateKey, err = marshaler.Marshal(*keyPair)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal RSA key pair: %w", err)
		}
	case ECC:
		generator := &crypto.ECCGenerator{}
		keyPair, err := generator.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECC key pair: %w", err)
		}

		marshaler := crypto.NewECCMarshaler()
		publicKey, privateKey, err = marshaler.Marshal(*keyPair)
		if err != nil {
			return nil, fmt.Errorf("failed to encode ECC key pair: %w", err)
		}
	}

	lastSignature := base64.StdEncoding.EncodeToString([]byte(id))

	return &SignatureDevice{
		ID:               id,
		Label:            label,
		Algorithm:        algorithm,
		SignatureCounter: 0,
		LastSignature:    lastSignature,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
	}, nil
}

func (d *SignatureDevice) GetSigner() (crypto.Signer, error) {
	switch d.Algorithm {
	case RSA:
		marshaler := crypto.NewRSAMarshaler()
		keyPair, err := marshaler.Unmarshal(d.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal RSA key pair: %w", err)
		}
		return &crypto.RSASigner{PrivateKey: keyPair.Private}, nil
	case ECC:
		marshaler := crypto.NewECCMarshaler()
		keyPair, err := marshaler.Unmarshal(d.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal ECC key pair: %w", err)
		}
		return &crypto.ECCSigner{PrivateKey: keyPair.Private}, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", d.Algorithm)
	}
}

func (d *SignatureDevice) SignTransaction(data string) (string, string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	securedData := fmt.Sprintf("%d_%s_%s", d.SignatureCounter, data, d.LastSignature)

	signer, err := d.GetSigner()
	if err != nil {
		return "", "", fmt.Errorf("failed to get signer: %w", err)
	}

	signature, err := signer.Sign([]byte(securedData))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign data: %w", err)
	}

	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	d.LastSignature = encodedSignature
	d.SignatureCounter++

	return encodedSignature, securedData, nil
}

func ParseSecuredData(securedData string) (int, string, string, error) {
	parts := strings.SplitN(securedData, "_", 3)
	if len(parts) != 3 {
		return 0, "", "", errors.New("invalid secured data format")
	}

	counter, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid signature counter: %w", err)
	}

	return counter, parts[1], parts[2], nil
}

func ValidateID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid UUID format")
	}
	return nil
}

func (d *SignatureDevice) Clone() *SignatureDevice {
	if d == nil {
		return nil
	}

	clone := &SignatureDevice{
		ID:               d.ID,
		Label:            d.Label,
		Algorithm:        d.Algorithm,
		SignatureCounter: d.SignatureCounter,
		LastSignature:    d.LastSignature,
	}

	if d.PublicKey != nil {
		clone.PublicKey = make([]byte, len(d.PublicKey))
		copy(clone.PublicKey, d.PublicKey)
	}

	if d.PrivateKey != nil {
		clone.PrivateKey = make([]byte, len(d.PrivateKey))
		copy(clone.PrivateKey, d.PrivateKey)
	}

	return clone
}
