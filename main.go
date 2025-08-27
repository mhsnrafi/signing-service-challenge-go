package main

import (
	"log"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/api"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
)

const (
	ListenAddress = ":8080"
)

func main() {
	repository := persistence.NewInMemoryDeviceRepository()

	deviceHandler := api.NewDeviceHandler(repository)
	
	server := api.NewServer(ListenAddress, deviceHandler)

	log.Printf("Starting server on %s", ListenAddress)
	if err := server.Run(); err != nil {
		log.Fatalf("Could not start server on %s: %v", ListenAddress, err)
	}
}
