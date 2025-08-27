# Signature Service

A RESTful API service that allows customers to create signature devices and sign transaction data.

## Overview

The Signature Service manages signature devices that can sign arbitrary transaction data. Each device has a unique identifier, a signature algorithm (RSA or ECDSA), a label, and a signature counter that tracks how many signatures have been created with the device.

When signing transaction data, the service extends the raw data with the current signature counter and the last signature to increase security. The resulting string is signed using the device's private key, and the signature is returned to the client along with the signed data.

## Features

- Create signature devices with RSA or ECDSA algorithms
- Sign transaction data with a signature device
- List all signature devices
- Retrieve a signature device by ID
- Thread-safe operations for concurrent access
- In-memory storage with future database migration in mind

## API Endpoints

### Create a Signature Device

```
POST /api/v0/devices
```

Request body:
```json
{
  "id": "optional-uuid",
  "algorithm": "RSA|ECC",
  "label": "My Device"
}
```

If the `id` field is not provided, a new UUID will be generated.

Response:
```json
{
  "data": {
    "id": "device-uuid",
    "label": "My Device",
    "algorithm": "RSA",
    "signature_counter": 0,
    "public_key": "base64-encoded-public-key"
  }
}
```

### List All Signature Devices

```
GET /api/v0/devices
```

Response:
```json
{
  "data": [
    {
      "id": "device-uuid-1",
      "label": "Device 1",
      "algorithm": "RSA",
      "signature_counter": 0,
      "public_key": "base64-encoded-public-key"
    },
    {
      "id": "device-uuid-2",
      "label": "Device 2",
      "algorithm": "ECC",
      "signature_counter": 0,
      "public_key": "base64-encoded-public-key"
    }
  ]
}
```

### Get a Signature Device

```
GET /api/v0/devices/{device-id}
```

Response:
```json
{
  "data": {
    "id": "device-uuid",
    "label": "My Device",
    "algorithm": "RSA",
    "signature_counter": 0,
    "public_key": "base64-encoded-public-key"
  }
}
```

### Sign a Transaction

```
POST /api/v0/devices/{device-id}/sign
```

Request body:
```json
{
  "data": "data-to-be-signed"
}
```

Response:
```json
{
  "data": {
    "signature": "base64-encoded-signature",
    "signed_data": "0_data-to-be-signed_base64-encoded-device-id"
  }
}
```

## Running the Service

1. Make sure you have Go 1.20 or later installed.
2. Clone the repository.
3. Run the service:

```bash
go run main.go
```

The service will start on port 8080.

## Testing

The service includes comprehensive tests for the domain model, storage layer, and API endpoints. To run the tests:

```bash
go test ./...
```

## Design Decisions

### Domain Model

The domain model is centered around the `SignatureDevice` struct, which encapsulates all the properties and behaviors of a signature device. The device is responsible for generating and storing its key pair, tracking its signature counter, and signing transaction data.

### Storage Layer

The storage layer is designed with a repository interface that can be implemented by different storage backends. Currently, an in-memory implementation is provided, but it would be easy to add a database implementation in the future.

### API Layer

The API layer follows RESTful principles and provides endpoints for creating, retrieving, and using signature devices. The API is designed to be easy to use and understand, with clear request and response formats.

### Concurrency

The service is designed to handle concurrent access to signature devices. The in-memory repository uses a read-write mutex to ensure thread safety, and the signature device itself uses a mutex to ensure that signature operations are atomic.

### Extensibility

The service is designed to be easily extended with new signature algorithms. The `Signer` interface abstracts away the details of the signature algorithm, allowing new algorithms to be added without changing the core domain logic.

## AI Tools Used

This project was implemented with the assistance of GitHub Copilot, which was used for code completion and suggestions throughout the development process.
