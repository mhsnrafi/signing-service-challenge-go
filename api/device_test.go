package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
	"github.com/google/uuid"
)

func TestCreateDevice(t *testing.T) {
	repo := persistence.NewInMemoryDeviceRepository()
	handler := NewDeviceHandler(repo)

	validID := uuid.New().String()
	validRequest := CreateDeviceRequest{
		ID:        validID,
		Algorithm: "RSA",
		Label:     "Test Device",
	}
	requestBody, _ := json.Marshal(validRequest)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()

	handler.CreateDevice(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response Response
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Failed to convert response data to map")
	}

	if responseData["id"] != validID {
		t.Errorf("Expected device ID to be %s, got %s", validID, responseData["id"])
	}

	if responseData["label"] != "Test Device" {
		t.Errorf("Expected device label to be 'Test Device', got %s", responseData["label"])
	}

	if responseData["algorithm"] != "RSA" {
		t.Errorf("Expected device algorithm to be 'RSA', got %s", responseData["algorithm"])
	}

	invalidRequest := CreateDeviceRequest{
		ID:        uuid.New().String(),
		Algorithm: "INVALID",
		Label:     "Test Device",
	}
	requestBody, _ = json.Marshal(invalidRequest)
	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices", bytes.NewBuffer(requestBody))
	rr = httptest.NewRecorder()

	handler.CreateDevice(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v0/devices", nil)
	rr = httptest.NewRecorder()

	handler.CreateDevice(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestGetDevice(t *testing.T) {
	repo := persistence.NewInMemoryDeviceRepository()
	handler := NewDeviceHandler(repo)

	id := uuid.New().String()
	device, err := domain.NewSignatureDevice(id, domain.RSA, "Test Device")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	err = repo.Create(context.Background(), device)
	if err != nil {
		t.Fatalf("Failed to create device in repository: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v0/devices/"+id, nil)
	rr := httptest.NewRecorder()
	req.URL.Path = "/api/v0/devices/" + id

	handler.GetDevice(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Failed to convert response data to map")
	}

	if responseData["id"] != id {
		t.Errorf("Expected device ID to be %s, got %s", id, responseData["id"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v0/devices/non-existent-id", nil)
	rr = httptest.NewRecorder()
	req.URL.Path = "/api/v0/devices/non-existent-id"

	handler.GetDevice(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices/"+id, nil)
	rr = httptest.NewRecorder()

	handler.GetDevice(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestListDevices(t *testing.T) {
	repo := persistence.NewInMemoryDeviceRepository()
	handler := NewDeviceHandler(repo)

	for i := 0; i < 3; i++ {
		id := uuid.New().String()
		device, err := domain.NewSignatureDevice(id, domain.RSA, "Test Device")
		if err != nil {
			t.Fatalf("Failed to create device: %v", err)
		}
		err = repo.Create(context.Background(), device)
		if err != nil {
			t.Fatalf("Failed to create device in repository: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v0/devices", nil)
	rr := httptest.NewRecorder()

	handler.ListDevices(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response Response
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	responseData, ok := response.Data.([]interface{})
	if !ok {
		t.Fatalf("Failed to convert response data to slice")
	}

	if len(responseData) != 3 {
		t.Errorf("Expected 3 devices, got %d", len(responseData))
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices", nil)
	rr = httptest.NewRecorder()

	handler.ListDevices(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestSignTransaction(t *testing.T) {
	repo := persistence.NewInMemoryDeviceRepository()
	handler := NewDeviceHandler(repo)

	id := uuid.New().String()
	device, err := domain.NewSignatureDevice(id, domain.RSA, "Test Device")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	err = repo.Create(context.Background(), device)
	if err != nil {
		t.Fatalf("Failed to create device in repository: %v", err)
	}

	validRequest := SignTransactionRequest{Data: "test data"}
	requestBody, _ := json.Marshal(validRequest)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/devices/"+id+"/sign", bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()
	req.URL.Path = "/api/v0/devices/" + id + "/sign"

	handler.SignTransaction(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Failed to convert response data to map")
	}

	if responseData["signature"] == "" {
		t.Errorf("Expected signature to be non-empty")
	}

	if responseData["signed_data"] == "" {
		t.Errorf("Expected signed data to be non-empty")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices/non-existent-id/sign", bytes.NewBuffer(requestBody))
	rr = httptest.NewRecorder()
	req.URL.Path = "/api/v0/devices/non-existent-id/sign"

	handler.SignTransaction(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	trailingSlashRequest := SignTransactionRequest{Data: "test data with trailing slash"}
	trailingSlashRequestBody, _ := json.Marshal(trailingSlashRequest)
	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices/"+id+"/sign/", bytes.NewBuffer(trailingSlashRequestBody))
	rr = httptest.NewRecorder()
	req.URL.Path = "/api/v0/devices/" + id + "/sign/"

	handler.SignTransaction(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code for URL with trailing slash: got %v want %v", status, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v0/devices/"+id+"/sign", nil)
	rr = httptest.NewRecorder()

	handler.SignTransaction(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleDeviceRequests(t *testing.T) {
	repo := persistence.NewInMemoryDeviceRepository()
	handler := NewDeviceHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v0/devices", nil)
	rr := httptest.NewRecorder()

	handler.HandleDeviceRequests(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %s", contentType)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v0/devices", nil)
	rr = httptest.NewRecorder()

	handler.HandleDeviceRequests(rr, req)

	contentType = rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %s", contentType)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v0/devices/some-id", nil)
	rr = httptest.NewRecorder()

	handler.HandleDeviceRequests(rr, req)

	contentType = rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %s", contentType)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices/some-id/sign", nil)
	rr = httptest.NewRecorder()

	handler.HandleDeviceRequests(rr, req)

	contentType = rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %s", contentType)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v0/devices/some-id/sign/", nil)
	rr = httptest.NewRecorder()

	handler.HandleDeviceRequests(rr, req)

	contentType = rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got %s", contentType)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v0/non-existent", nil)
	rr = httptest.NewRecorder()

	handler.HandleDeviceRequests(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
