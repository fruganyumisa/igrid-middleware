package main

import (
	"context"
	"log"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware/internal/models"
)

func main() {
	log.Println("🔧 Proper Modbus Package Usage Demo")
	log.Println("===================================")

	// ✅ CORRECT: Create modbus server with just device ID
	server, err := modbus.NewServer("SMART_METER_001")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// ✅ OPTIONAL: Add HTTP callback for DMS communication
	httpCallback := modbus.CreateHTTPCallback("http://localhost:8080", 30*time.Second)
	server.SetHTTPCallback(httpCallback)

	// ✅ OPTIONAL: Add custom data processing callback
	server.SetDataCallback(func(data *models.SmartGridData) error {
		log.Printf("🏭 Custom Processing: Device %s has voltage %.2fV, current %.3fA",
			data.DeviceID, data.Voltage, data.Current)

		// Could add custom logic here like:
		// - Alarms if voltage too high/low
		// - Data logging to database
		// - Custom calculations

		return nil
	})

	// Start the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("🌐 Starting Modbus server on :5020...")
		if err := server.Start(ctx, ":5020"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Let it run for a few seconds
	time.Sleep(3 * time.Second)

	log.Println("✅ Modbus server running with:")
	log.Println("   - Device ID: SMART_METER_001")
	log.Println("   - HTTP DMS integration: Configured")
	log.Println("   - Custom data processing: Enabled")
	log.Println("   - Smart grid data: Ready for reading")
	log.Println("")
	log.Println("📡 Connect a Modbus client to see the magic happen!")
	log.Println("   Read registers 0-14 to trigger complete data processing")

	// Demonstrate direct HTTP functionality
	log.Println("")
	log.Println("🚀 Testing direct HTTP functionality...")

	testPayload := &models.DMSPayload{
		DeviceID:    "SMART_METER_001",
		Voltage:     240.5,
		Current:     12.3,
		Temperature: 25.8,
	}

	// This would work if we had a real endpoint
	// err = modbus.SendHTTPRequest("http://httpbin.org/post", testPayload, 10*time.Second)
	log.Printf("Would send: %+v", testPayload)

	log.Println("Demo complete! 🎉")
}
