package persistence

import (
	"context"
	"testing"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/google/uuid"
)

func TestInMemoryDeviceRepository(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	id := uuid.New().String()
	label := "Test Device"
	device, err := domain.NewSignatureDevice(id, domain.RSA, label)
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}

	err = repo.Create(ctx, device)
	if err != nil {
		t.Fatalf("Failed to create device in repository: %v", err)
	}

	retrievedDevice, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get device from repository: %v", err)
	}
	if retrievedDevice.ID != device.ID {
		t.Errorf("Expected device ID to be %s, got %s", device.ID, retrievedDevice.ID)
	}
	if retrievedDevice.Label != device.Label {
		t.Errorf("Expected device label to be %s, got %s", device.Label, retrievedDevice.Label)
	}
	if retrievedDevice.Algorithm != device.Algorithm {
		t.Errorf("Expected device algorithm to be %s, got %s", device.Algorithm, retrievedDevice.Algorithm)
	}

	devices, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list devices from repository: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(devices))
	}
	if devices[0].ID != device.ID {
		t.Errorf("Expected device ID to be %s, got %s", device.ID, devices[0].ID)
	}

	device.Label = "Updated Label"
	err = repo.Update(ctx, device)
	if err != nil {
		t.Fatalf("Failed to update device in repository: %v", err)
	}
	updatedDevice, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get updated device from repository: %v", err)
	}
	if updatedDevice.Label != "Updated Label" {
		t.Errorf("Expected device label to be 'Updated Label', got %s", updatedDevice.Label)
	}

	err = repo.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete device from repository: %v", err)
	}
	_, err = repo.Get(ctx, id)
	if err == nil {
		t.Errorf("Expected error when getting deleted device")
	}
	devices, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list devices from repository: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("Expected 0 devices, got %d", len(devices))
	}
}

func TestInMemoryDeviceRepositoryErrors(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	err := repo.Create(ctx, nil)
	if err == nil {
		t.Errorf("Expected error when creating nil device")
	}

	device := &domain.SignatureDevice{
		ID:    "",
		Label: "Test Device",
	}
	err = repo.Create(ctx, device)
	if err == nil {
		t.Errorf("Expected error when creating device with empty ID")
	}

	id := uuid.New().String()
	validDevice, err := domain.NewSignatureDevice(id, domain.RSA, "Test Device")
	if err != nil {
		t.Fatalf("Failed to create device: %v", err)
	}
	err = repo.Create(ctx, validDevice)
	if err != nil {
		t.Fatalf("Failed to create device in repository: %v", err)
	}

	duplicateDevice, err := domain.NewSignatureDevice(id, domain.ECC, "Duplicate Device")
	if err != nil {
		t.Fatalf("Failed to create duplicate device: %v", err)
	}
	err = repo.Create(ctx, duplicateDevice)
	if err == nil {
		t.Errorf("Expected error when creating device with duplicate ID")
	}

	_, err = repo.Get(ctx, "non-existent-id")
	if err == nil {
		t.Errorf("Expected error when getting non-existent device")
	}

	err = repo.Update(ctx, nil)
	if err == nil {
		t.Errorf("Expected error when updating nil device")
	}

	err = repo.Update(ctx, &domain.SignatureDevice{
		ID:    "",
		Label: "Test Device",
	})
	if err == nil {
		t.Errorf("Expected error when updating device with empty ID")
	}

	err = repo.Update(ctx, &domain.SignatureDevice{
		ID:    "non-existent-id",
		Label: "Test Device",
	})
	if err == nil {
		t.Errorf("Expected error when updating non-existent device")
	}

	err = repo.Delete(ctx, "non-existent-id")
	if err == nil {
		t.Errorf("Expected error when deleting non-existent device")
	}
}

func TestInMemoryDeviceRepositoryConcurrency(t *testing.T) {
	repo := NewInMemoryDeviceRepository()
	ctx := context.Background()

	const numDevices = 10
	done := make(chan bool, numDevices)

	for i := 0; i < numDevices; i++ {
		go func() {
			goCtx := context.Background()

			id := uuid.New().String()
			device, err := domain.NewSignatureDevice(id, domain.RSA, "Concurrent Device")
			if err != nil {
				t.Errorf("Failed to create device: %v", err)
				done <- false
				return
			}

			err = repo.Create(goCtx, device)
			if err != nil {
				t.Errorf("Failed to create device in repository: %v", err)
				done <- false
				return
			}

			done <- true
		}()
	}

	for i := 0; i < numDevices; i++ {
		success := <-done
		if !success {
			t.Fatalf("Concurrent device creation failed")
		}
	}

	devices, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list devices from repository: %v", err)
	}
	if len(devices) != numDevices {
		t.Errorf("Expected %d devices, got %d", numDevices, len(devices))
	}
}
