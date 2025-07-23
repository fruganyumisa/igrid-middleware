package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware/internal/models"
)

// MockDMSServer creates a simple HTTP server to simulate the DMS application
func startMockDMSServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/devices/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log received data
		log.Printf("DMS Server received data: %+v", payload)

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "message": "Data received successfully"}`))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Mock DMS server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("DMS server error: %v", err)
		}
	}()

	return server
}

func main() {
	log.Println("Starting iGrid Middleware with Smart Grid Data Processing...")

	// Start mock DMS server
	dmsServer := startMockDMSServer()
	defer dmsServer.Close()

	// Wait a moment for DMS server to start
	time.Sleep(1 * time.Second)

	// Create logger (removed as we don't need the complex logger interface)
	// logger := &MockLogger{}

	// Create Modbus server with smart grid data processing
	modbusServer, err := modbus.NewServer("SMART_GRID_DEVICE_001")
	if err != nil {
		log.Fatalf("Failed to create Modbus server: %v", err)
	}

	// Set up HTTP callback for DMS communication
	httpCallback := modbus.CreateHTTPCallback("http://localhost:8080", 30*time.Second)
	modbusServer.SetHTTPCallback(httpCallback)

	// Optional: Set up data callback for additional processing
	modbusServer.SetDataCallback(func(data *models.SmartGridData) error {
		log.Printf("📊 Additional processing for device %s: V=%.2f, I=%.3f, T=%.2f",
			data.DeviceID, data.Voltage, data.Current, data.Temperature)
		return nil
	})
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Modbus server
	go func() {
		log.Println("Starting Modbus server on :5020...")
		if err := modbusServer.Start(ctx, ":5020"); err != nil {
			log.Printf("Modbus server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("iGrid Middleware is running...")
	log.Println("- Modbus server listening on :5020")
	log.Println("- Mock DMS server running on :8080")
	log.Println("- Device ID: SMART_GRID_DEVICE_001")
	log.Println("Connect a Modbus client to read registers 0-14 to see data routing in action!")

	<-sigChan
	log.Println("Shutting down...")

	// Graceful shutdown
	cancel()

	// Shutdown DMS server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	dmsServer.Shutdown(shutdownCtx)

	// Stop Modbus server
	modbusServer.Stop()

	log.Println("Shutdown complete.")
}
