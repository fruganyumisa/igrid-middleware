package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	log.Println("🔧 Modbus Client - Testing Smart Grid Data Processing")
	log.Println("====================================================")

	// Create Modbus client
	handler := modbus.NewTCPClientHandler("localhost:5020")
	handler.Timeout = 10 * time.Second

	err := handler.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Modbus server: %v", err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	log.Println("✅ Connected to Modbus server on localhost:5020")
	log.Println("")

	// Read all smart grid registers (0-14) to trigger data processing
	log.Println("📊 Reading smart grid registers 0-14...")

	// Read holding registers 0-14 (15 registers total)
	results, err := client.ReadHoldingRegisters(0, 15)
	if err != nil {
		log.Fatalf("❌ Failed to read holding registers: %v", err)
	}

	log.Printf("✅ Successfully read %d bytes from registers 0-14", len(results))
	log.Println("")

	// Convert bytes to uint16 values and display
	log.Println("📈 Raw Register Values:")
	for i := 0; i < len(results); i += 2 {
		if i+1 < len(results) {
			value := uint16(results[i])<<8 | uint16(results[i+1])
			regAddr := i / 2

			// Show interpretation based on our smart grid mapping
			interpretation := getSmartGridInterpretation(regAddr, value)
			log.Printf("   Register %2d: %5d (0x%04X) - %s", regAddr, value, value, interpretation)
		}
	}

	log.Println("")
	log.Println("🚀 Check the server logs to see the data processing and DMS routing!")
	log.Println("   The server should have detected a complete dataset and generated DMS payload.")
}

func getSmartGridInterpretation(address int, value uint16) string {
	switch address {
	case 0:
		voltage := float64(value) / 100.0
		return fmt.Sprintf("Voltage: %.2fV", voltage)
	case 1:
		current := float64(value) / 1000.0
		return fmt.Sprintf("Current: %.3fA", current)
	case 2:
		power := float64(value) / 10.0
		return fmt.Sprintf("Active Power: %.1fW", power)
	case 3:
		reactivePower := float64(value) / 100.0
		return fmt.Sprintf("Reactive Power: %.2fVAR", reactivePower)
	case 4:
		frequency := float64(value) / 100.0
		return fmt.Sprintf("Frequency: %.2fHz", frequency)
	case 5:
		status := "Open"
		if value == 1 {
			status = "Closed"
		}
		return fmt.Sprintf("Breaker Status: %s", status)
	case 6:
		return fmt.Sprintf("Tap Position: %d", value)
	case 7:
		return "Energy High Word"
	case 8:
		return "Energy Low Word"
	case 9:
		temp := float64(value) / 100.0
		return fmt.Sprintf("Temperature: %.2f°C", temp)
	case 10:
		humidity := float64(value) / 100.0
		return fmt.Sprintf("Humidity: %.2f%%", humidity)
	case 11:
		alarmStatus := "No Alarm"
		if value != 0 {
			alarmStatus = "Alarm Active"
		}
		return fmt.Sprintf("Alarm: %s", alarmStatus)
	case 12:
		return "Timestamp High Word"
	case 13:
		return "Timestamp Low Word"
	case 14:
		return "Device ID"
	default:
		return "Unknown Register"
	}
}
