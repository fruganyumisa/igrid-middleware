package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware/internal/models"
)

// MockRouter implements DataCallback for testing
type MockRouter struct{}

func (r *MockRouter) RouteSmartGridData(data *models.SmartGridData) error {
	// Convert to DMS format
	dmsPayload := models.DMSPayload{
		DeviceID:    data.DeviceID,
		Voltage:     data.Voltage,
		Current:     data.Current,
		Temperature: data.Temperature,
	}

	// Log the DMS payload
	payloadJSON, _ := json.Marshal(dmsPayload)
	log.Printf("🚀 DMS Payload: %s", string(payloadJSON))

	// Log full smart grid data for reference
	fullDataJSON, _ := json.Marshal(data)
	log.Printf("📊 Full Smart Grid Data: %s", string(fullDataJSON))

	return nil
}

func main() {
	log.Println("🔌 Smart Grid Data Processing Demo")
	log.Println("===============================")

	// Create mock router
	router := &MockRouter{}

	// Create Modbus server with smart grid data
	server, err := modbus.NewServer("SMART_GRID_METER_001")
	if err != nil {
		log.Fatalf("❌ Failed to create Modbus server: %v", err)
	}

	// Set up data callback to handle processed data
	server.SetDataCallback(router.RouteSmartGridData)

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("🌐 Starting Modbus server on :5020...")
		if err := server.Start(ctx, ":5020"); err != nil {
			log.Printf("Modbus server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	log.Println("✅ Modbus server is running with smart grid data:")
	log.Println("   - Device ID: SMART_GRID_METER_001")
	log.Println("   - Voltage: 262.47V (register 0)")
	log.Println("   - Current: 14.56A (register 1)")
	log.Println("   - Active Power: 1521.0W (register 2)")
	log.Println("   - Temperature: 63.70°C (register 9)")
	log.Println("")
	log.Println("📡 Connect a Modbus client to read registers 0-14 to see:")
	log.Println("   1. Human-readable data interpretation")
	log.Println("   2. Automatic routing to DMS format")
	log.Println("   3. JSON payload generation")
	log.Println("")
	log.Println("Example: Use a Modbus client to read holding registers 0-14")
	log.Println("The server will automatically detect complete datasets and route them!")

	// Keep running
	select {}
}
