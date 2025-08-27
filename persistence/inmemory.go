package persistence

import (
	"context"
	"errors"
	"sync"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *domain.SignatureDevice) error
	Get(ctx context.Context, id string) (*domain.SignatureDevice, error)
	List(ctx context.Context) ([]*domain.SignatureDevice, error)
	Update(ctx context.Context, device *domain.SignatureDevice) error
	Delete(ctx context.Context, id string) error
}

type InMemoryDeviceRepository struct {
	devices map[string]*domain.SignatureDevice
	mu      sync.RWMutex
}

func NewInMemoryDeviceRepository() *InMemoryDeviceRepository {
	return &InMemoryDeviceRepository{
		devices: make(map[string]*domain.SignatureDevice),
	}
}

func (r *InMemoryDeviceRepository) Create(ctx context.Context, device *domain.SignatureDevice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if device == nil {
		return errors.New("device cannot be nil")
	}
	if device.ID == "" {
		return errors.New("device ID cannot be empty")
	}
	if _, exists := r.devices[device.ID]; exists {
		return errors.New("device with this ID already exists")
	}

	r.devices[device.ID] = device.Clone()
	return nil
}

func (r *InMemoryDeviceRepository) Get(ctx context.Context, id string) (*domain.SignatureDevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, exists := r.devices[id]
	if !exists {
		return nil, errors.New("device not found")
	}
	return device.Clone(), nil
}

func (r *InMemoryDeviceRepository) List(ctx context.Context) ([]*domain.SignatureDevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	devices := make([]*domain.SignatureDevice, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device.Clone())
	}
	return devices, nil
}

func (r *InMemoryDeviceRepository) Update(ctx context.Context, device *domain.SignatureDevice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if device == nil {
		return errors.New("device cannot be nil")
	}
	if device.ID == "" {
		return errors.New("device ID cannot be empty")
	}
	if _, exists := r.devices[device.ID]; !exists {
		return errors.New("device not found")
	}

	r.devices[device.ID] = device.Clone()
	return nil
}

func (r *InMemoryDeviceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.devices[id]; !exists {
		return errors.New("device not found")
	}
	delete(r.devices, id)
	return nil
}
