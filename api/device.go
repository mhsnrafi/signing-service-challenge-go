package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
	"github.com/google/uuid"
)

type CreateDeviceRequest struct {
	ID        string `json:"id"`
	Algorithm string `json:"algorithm"`
	Label     string `json:"label"`
}

type CreateDeviceResponse struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	Algorithm        string `json:"algorithm"`
	SignatureCounter int    `json:"signature_counter"`
	PublicKey        []byte `json:"public_key"`
}

type SignTransactionRequest struct {
	Data string `json:"data"`
}

type SignTransactionResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

type DeviceHandler struct {
	repository persistence.DeviceRepository
}

func NewDeviceHandler(repository persistence.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{repository: repository}
}

func (h *DeviceHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{http.StatusText(http.StatusMethodNotAllowed)})
		return
	}

	var request CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if request.ID == "" {
		request.ID = uuid.New().String()
	} else {
		if err := domain.ValidateID(request.ID); err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, []string{err.Error()})
			return
		}
	}

	algorithm := domain.SignatureAlgorithm(strings.ToUpper(request.Algorithm))
	if algorithm != domain.RSA && algorithm != domain.ECC {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"Invalid algorithm. Supported algorithms: RSA, ECC"})
		return
	}

	device, err := domain.NewSignatureDevice(request.ID, algorithm, request.Label)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{fmt.Sprintf("Failed to create signature device: %v", err)})
		return
	}

	if err := h.repository.Create(r.Context(), device); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{fmt.Sprintf("Failed to store signature device: %v", err)})
		return
	}

	response := CreateDeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        string(device.Algorithm),
		SignatureCounter: device.SignatureCounter,
		PublicKey:        device.PublicKey,
	}

	WriteAPIResponse(w, http.StatusCreated, response)
}

func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{http.StatusText(http.StatusMethodNotAllowed)})
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"Invalid URL path"})
		return
	}
	id := parts[len(parts)-1]

	device, err := h.repository.Get(r.Context(), id)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, []string{"Device not found"})
		return
	}

	response := CreateDeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        string(device.Algorithm),
		SignatureCounter: device.SignatureCounter,
		PublicKey:        device.PublicKey,
	}

	WriteAPIResponse(w, http.StatusOK, response)
}

func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{http.StatusText(http.StatusMethodNotAllowed)})
		return
	}

	devices, err := h.repository.List(r.Context())
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{fmt.Sprintf("Failed to retrieve devices: %v", err)})
		return
	}

	response := make([]CreateDeviceResponse, 0, len(devices))
	for _, device := range devices {
		response = append(response, CreateDeviceResponse{
			ID:               device.ID,
			Label:            device.Label,
			Algorithm:        string(device.Algorithm),
			SignatureCounter: device.SignatureCounter,
			PublicKey:        device.PublicKey,
		})
	}

	WriteAPIResponse(w, http.StatusOK, response)
}

func (h *DeviceHandler) SignTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{http.StatusText(http.StatusMethodNotAllowed)})
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	var filteredParts []string
	for _, part := range parts {
		if part != "" {
			filteredParts = append(filteredParts, part)
		}
	}

	if len(filteredParts) < 5 || filteredParts[len(filteredParts)-1] != "sign" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"Invalid URL path"})
		return
	}
	id := filteredParts[len(filteredParts)-2]

	device, err := h.repository.Get(r.Context(), id)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, []string{"Device not found"})
		return
	}

	var request SignTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	signature, signedData, err := device.SignTransaction(request.Data)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{fmt.Sprintf("Failed to sign transaction: %v", err)})
		return
	}

	if err := h.repository.Update(r.Context(), device); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{fmt.Sprintf("Failed to update device: %v", err)})
		return
	}

	response := SignTransactionResponse{
		Signature:  signature,
		SignedData: signedData,
	}

	WriteAPIResponse(w, http.StatusOK, response)
}

func (h *DeviceHandler) HandleDeviceRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	w.Header().Set("Content-Type", "application/json")

	if path == "/api/v0/devices" && r.Method == http.MethodPost {
		h.CreateDevice(w, r)
	} else if path == "/api/v0/devices" && r.Method == http.MethodGet {
		h.ListDevices(w, r)
	} else if strings.HasPrefix(path, "/api/v0/devices/") && !strings.Contains(path, "/sign") && r.Method == http.MethodGet {
		h.GetDevice(w, r)
	} else if (strings.HasSuffix(path, "/sign") || strings.HasSuffix(path, "/sign/")) && r.Method == http.MethodPost {
		h.SignTransaction(w, r)
	} else {
		WriteErrorResponse(w, http.StatusNotFound, []string{"Endpoint not found"})
	}
}
