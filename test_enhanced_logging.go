package main

import (
	"context"
	"log"
	"time"

	modbusserver "github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	modbusclient "github.com/goburrow/modbus"
)

func main() {
	log.Println("=== Testing Enhanced Modbus Logging ===")

	// Start server on port 5020 (no root required)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := modbusserver.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		if err := server.Start(ctx, ":5020"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)
	log.Println("Server started on port 5020")

	// Create client and test different operations
	handler := modbusclient.NewTCPClientHandler("localhost:5020")
	handler.Timeout = 5 * time.Second
	client := modbusclient.NewClient(handler)

	err = handler.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer handler.Close()

	log.Println("Connected to server. Testing enhanced logging...")

	// Test 1: Read holding registers (should show voltage interpretation)
	log.Println("1. Reading holding register 0 (voltage)...")
	_, err = client.ReadHoldingRegisters(0, 1)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Test 2: Write single register with a voltage value
	log.Println("2. Writing voltage value 2400 to register 0...")
	_, err = client.WriteSingleRegister(0, 2400) // 240.0V when scaled by /10
	if err != nil {
		log.Printf("Error: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Test 3: Write current value to register 1
	log.Println("3. Writing current value 1500 to register 1...")
	_, err = client.WriteSingleRegister(1, 1500) // 15.00A when scaled by /100
	if err != nil {
		log.Printf("Error: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Test 4: Write multiple registers
	log.Println("4. Writing multiple register values...")
	_, err = client.WriteMultipleRegisters(2, 3, []byte{0x13, 0x88, 0x27, 0x10, 0x03, 0xE8})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	time.Sleep(1 * time.Second)

	log.Println("Test completed! Check the logs above for enhanced human-readable decoding.")

	// Stop server
	cancel()
	time.Sleep(1 * time.Second)
}
