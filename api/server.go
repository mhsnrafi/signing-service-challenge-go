package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Response is the generic API response container.
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	listenAddress string
	deviceHandler *DeviceHandler
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string, deviceHandler *DeviceHandler) *Server {
	return &Server{
		listenAddress: listenAddress,
		deviceHandler: deviceHandler,
	}
}

// Run registers all HandlerFuncs for the existing HTTP routes and starts the Server.
func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.Handle("/api/v0/health", http.HandlerFunc(s.Health))
	mux.HandleFunc("/api/v0/devices", s.deviceHandler.HandleDeviceRequests)
	mux.HandleFunc("/api/v0/devices/", s.deviceHandler.HandleDeviceRequests)

	server := &http.Server{
		Addr:    s.listenAddress,
		Handler: mux,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine so that it doesn't block.
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	// Listen for an interrupt signal from the OS.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error from the server.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-osSignals:
		// Create a context with a timeout to allow for graceful shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server.
		if err := server.Shutdown(ctx); err != nil {
			// If shutdown fails, force close.
			if err := server.Close(); err != nil {
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
		}
	}

	return nil
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}
