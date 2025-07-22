package main

import (
	"fmt"
	"log"
	"time"

	"github.com/goburrow/modbus"
)

func main() {
	fmt.Println("Testing Modbus server logging...")

	// Create a Modbus TCP client
	handler := modbus.NewTCPClientHandler("localhost:502")
	handler.Timeout = 10 * time.Second
	client := modbus.NewClient(handler)

	err := handler.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer handler.Close()

	fmt.Println("Connected to Modbus server. Making test requests...")

	// Test 1: Read holding registers
	fmt.Println("1. Reading holding registers (0-2)...")
	results, err := client.ReadHoldingRegisters(0, 3)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Success: %02x\n", results)
	}

	time.Sleep(1 * time.Second)

	// Test 2: Write single register
	fmt.Println("2. Writing single register (address 0, value 9999)...")
	results, err = client.WriteSingleRegister(0, 9999)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Success: %02x\n", results)
	}

	time.Sleep(1 * time.Second)

	// Test 3: Read input registers
	fmt.Println("3. Reading input registers (0-1)...")
	results, err = client.ReadInputRegisters(0, 2)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Success: %02x\n", results)
	}

	time.Sleep(1 * time.Second)

	// Test 4: Write multiple registers
	fmt.Println("4. Writing multiple registers...")
	results, err = client.WriteMultipleRegisters(1, 2, []byte{0x12, 0x34, 0x56, 0x78})
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Success: %02x\n", results)
	}

	fmt.Println("\nTest completed! Check server logs for detailed request logging.")
}
