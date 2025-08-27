package domain

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewSignatureDevice(t *testing.T) {
	id := uuid.New().String()
	label := "Test Device"

	rsaDevice, err := NewSignatureDevice(id, RSA, label)
	if err != nil {
		t.Fatalf("Failed to create RSA device: %v", err)
	}
	if rsaDevice.ID != id {
		t.Errorf("Expected device ID to be %s, got %s", id, rsaDevice.ID)
	}
	if rsaDevice.Label != label {
		t.Errorf("Expected device label to be %s, got %s", label, rsaDevice.Label)
	}
	if rsaDevice.Algorithm != RSA {
		t.Errorf("Expected device algorithm to be %s, got %s", RSA, rsaDevice.Algorithm)
	}
	if rsaDevice.SignatureCounter != 0 {
		t.Errorf("Expected signature counter to be 0, got %d", rsaDevice.SignatureCounter)
	}
	if rsaDevice.LastSignature != base64.StdEncoding.EncodeToString([]byte(id)) {
		t.Errorf("Expected last signature to be base64-encoded device ID")
	}
	if len(rsaDevice.PublicKey) == 0 {
		t.Errorf("Expected public key to be non-empty")
	}
	if len(rsaDevice.PrivateKey) == 0 {
		t.Errorf("Expected private key to be non-empty")
	}

	eccDevice, err := NewSignatureDevice(id, ECC, label)
	if err != nil {
		t.Fatalf("Failed to create ECC device: %v", err)
	}
	if eccDevice.ID != id {
		t.Errorf("Expected device ID to be %s, got %s", id, eccDevice.ID)
	}
	if eccDevice.Label != label {
		t.Errorf("Expected device label to be %s, got %s", label, eccDevice.Label)
	}
	if eccDevice.Algorithm != ECC {
		t.Errorf("Expected device algorithm to be %s, got %s", ECC, eccDevice.Algorithm)
	}
	if eccDevice.SignatureCounter != 0 {
		t.Errorf("Expected signature counter to be 0, got %d", eccDevice.SignatureCounter)
	}
	if eccDevice.LastSignature != base64.StdEncoding.EncodeToString([]byte(id)) {
		t.Errorf("Expected last signature to be base64-encoded device ID")
	}
	if len(eccDevice.PublicKey) == 0 {
		t.Errorf("Expected public key to be non-empty")
	}
	if len(eccDevice.PrivateKey) == 0 {
		t.Errorf("Expected private key to be non-empty")
	}

	_, err = NewSignatureDevice("", RSA, label)
	if err == nil {
		t.Errorf("Expected error for empty device ID")
	}

	_, err = NewSignatureDevice(id, "INVALID", label)
	if err == nil {
		t.Errorf("Expected error for invalid algorithm")
	}
}

func TestSignTransaction(t *testing.T) {
	id := uuid.New().String()
	label := "Test Device"
	device, err := NewSignatureDevice(id, RSA, label)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	data := "test data"
	signature, signedData, err := device.SignTransaction(data)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}
	if signature == "" {
		t.Errorf("Expected signature to be non-empty")
	}

	counter, signedDataContent, lastSignature, err := ParseSecuredData(signedData)
	if err != nil {
		t.Fatalf("Failed to parse secured data: %v", err)
	}
	if counter != 0 {
		t.Errorf("Expected counter to be 0, got %d", counter)
	}
	if signedDataContent != data {
		t.Errorf("Expected signed data content to be %s, got %s", data, signedDataContent)
	}
	if lastSignature != base64.StdEncoding.EncodeToString([]byte(id)) {
		t.Errorf("Expected last signature to be base64-encoded device ID")
	}
	if device.SignatureCounter != 1 {
		t.Errorf("Expected signature counter to be 1, got %d", device.SignatureCounter)
	}
	if device.LastSignature != signature {
		t.Errorf("Expected last signature to be updated")
	}

	data2 := "test data 2"
	signature2, signedData2, err := device.SignTransaction(data2)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}
	if signature2 == "" {
		t.Errorf("Expected signature to be non-empty")
	}

	counter2, signedDataContent2, lastSignature2, err := ParseSecuredData(signedData2)
	if err != nil {
		t.Fatalf("Failed to parse secured data: %v", err)
	}
	if counter2 != 1 {
		t.Errorf("Expected counter to be 1, got %d", counter2)
	}
	if signedDataContent2 != data2 {
		t.Errorf("Expected signed data content to be %s, got %s", data2, signedDataContent2)
	}
	if lastSignature2 != signature {
		t.Errorf("Expected last signature to be the previous signature")
	}
	if device.SignatureCounter != 2 {
		t.Errorf("Expected signature counter to be 2, got %d", device.SignatureCounter)
	}
	if device.LastSignature != signature2 {
		t.Errorf("Expected last signature to be updated")
	}
}

func TestParseSecuredData(t *testing.T) {
	counter := 1
	data := "test data"
	lastSignature := "base64-encoded-signature"
	securedData := strings.Join([]string{"1", data, lastSignature}, "_")

	parsedCounter, parsedData, parsedLastSignature, err := ParseSecuredData(securedData)
	if err != nil {
		t.Fatalf("Failed to parse secured data: %v", err)
	}
	if parsedCounter != counter {
		t.Errorf("Expected counter to be %d, got %d", counter, parsedCounter)
	}
	if parsedData != data {
		t.Errorf("Expected data to be %s, got %s", data, parsedData)
	}
	if parsedLastSignature != lastSignature {
		t.Errorf("Expected last signature to be %s, got %s", lastSignature, parsedLastSignature)
	}

	invalidSecuredData := "invalid_format"
	_, _, _, err = ParseSecuredData(invalidSecuredData)
	if err == nil {
		t.Errorf("Expected error for invalid secured data format")
	}

	invalidSecuredData = "invalid_format_with_more_parts"
	_, _, _, err = ParseSecuredData(invalidSecuredData)
	if err == nil {
		t.Errorf("Expected error for invalid secured data format")
	}

	invalidSecuredData = "not-a-number_data_signature"
	_, _, _, err = ParseSecuredData(invalidSecuredData)
	if err == nil {
		t.Errorf("Expected error for invalid counter")
	}
}

func TestValidateID(t *testing.T) {
	validID := uuid.New().String()
	err := ValidateID(validID)
	if err != nil {
		t.Errorf("Expected no error for valid UUID, got %v", err)
	}

	invalidID := "not-a-uuid"
	err = ValidateID(invalidID)
	if err == nil {
		t.Errorf("Expected error for invalid UUID")
	}
}
