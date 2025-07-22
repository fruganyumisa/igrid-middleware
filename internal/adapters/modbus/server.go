package modbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tbrandon/mbserver"
)

type ModbusServer struct {
	server *mbserver.Server
	addr   string
}

type Message struct {
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
	Values    map[string]any `json:"values"`
}

func (m *ModbusServer) OnWrite(addr uint16, data []byte) {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	logEntry := map[string]interface{}{
		"timestamp":      timestamp,
		"protocol":       "modbus",
		"operation":      "write",
		"address":        addr,
		"data":           fmt.Sprintf("%02x", data),
		"server_address": m.addr,
	}

	logJSON, _ := json.Marshal(logEntry)
	log.Printf("Modbus Write Operation: %s", string(logJSON))
}

func NewServer() (*ModbusServer, error) {
	srv := mbserver.NewServer()

	modbusServer := &ModbusServer{
		server: srv,
		addr:   ":5020",
	}

	srv.HoldingRegisters[0] = 1234
	srv.HoldingRegisters[1] = 5678
	srv.HoldingRegisters[2] = 9012
	srv.InputRegisters[0] = 3456
	srv.InputRegisters[1] = 7890

	return modbusServer, nil
}

func getFunctionName(code uint8) string {
	functionNames := map[uint8]string{
		1:  "READ_COILS",
		2:  "READ_DISCRETE_INPUTS",
		3:  "READ_HOLDING_REGISTERS",
		4:  "READ_INPUT_REGISTERS",
		5:  "WRITE_SINGLE_COIL",
		6:  "WRITE_SINGLE_REGISTER",
		15: "WRITE_MULTIPLE_COILS",
		16: "WRITE_MULTIPLE_REGISTERS",
	}

	if name, exists := functionNames[code]; exists {
		return name
	}
	return fmt.Sprintf("UNKNOWN_FUNCTION_%d", code)
}

func (m *ModbusServer) Start(ctx context.Context, address string) error {
	m.addr = address

	log.Printf("Starting Modbus TCP server on %s with comprehensive request logging", address)

	// Enable debug mode for more verbose logging
	m.server.Debug = true

	// Register function handlers that will log requests
	m.server.RegisterFunctionHandler(1, m.logAndHandle(1, mbserver.ReadCoils))
	m.server.RegisterFunctionHandler(2, m.logAndHandle(2, mbserver.ReadDiscreteInputs))
	m.server.RegisterFunctionHandler(3, m.logAndHandle(3, mbserver.ReadHoldingRegisters))
	m.server.RegisterFunctionHandler(4, m.logAndHandle(4, mbserver.ReadInputRegisters))
	m.server.RegisterFunctionHandler(5, m.logAndHandle(5, mbserver.WriteSingleCoil))
	m.server.RegisterFunctionHandler(6, m.logAndHandle(6, mbserver.WriteHoldingRegister))
	m.server.RegisterFunctionHandler(15, m.logAndHandle(15, mbserver.WriteMultipleCoils))
	m.server.RegisterFunctionHandler(16, m.logAndHandle(16, mbserver.WriteHoldingRegisters))

	go func() {
		<-ctx.Done()
		log.Println("Shutting down Modbus server...")
		m.server.Close()
	}()

	// Start server and log the fact that it's listening
	log.Printf("Modbus server listening on %s - all requests will be logged", address)
	return m.server.ListenTCP(address)
}

// logAndHandle creates a wrapper function that logs requests before calling the original handler
func (m *ModbusServer) logAndHandle(funcCode uint8, originalHandler func(*mbserver.Server, mbserver.Framer) ([]byte, *mbserver.Exception)) func(*mbserver.Server, mbserver.Framer) ([]byte, *mbserver.Exception) {
	return func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
		// Log the request
		data := frame.GetData()
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)

		logEntry := map[string]interface{}{
			"timestamp":      timestamp,
			"protocol":       "modbus",
			"function_code":  funcCode,
			"function_name":  getFunctionName(funcCode),
			"data_length":    len(data),
			"raw_data":       fmt.Sprintf("%02x", data),
			"server_address": m.addr,
		}

		// Parse parameters based on function code
		switch funcCode {
		case 1, 2, 3, 4: // Read operations
			if len(data) >= 4 {
				startAddr := uint16(data[0])<<8 | uint16(data[1])
				quantity := uint16(data[2])<<8 | uint16(data[3])
				logEntry["start_address"] = startAddr
				logEntry["quantity"] = quantity
			}
		case 5, 6: // Write single operations
			if len(data) >= 4 {
				address := uint16(data[0])<<8 | uint16(data[1])
				value := uint16(data[2])<<8 | uint16(data[3])
				logEntry["address"] = address
				logEntry["value"] = value

				// Call OnWrite for write operations
				m.OnWrite(address, data[2:4])
			}
		case 15, 16: // Write multiple operations
			if len(data) >= 5 {
				startAddr := uint16(data[0])<<8 | uint16(data[1])
				quantity := uint16(data[2])<<8 | uint16(data[3])
				byteCount := data[4]
				logEntry["start_address"] = startAddr
				logEntry["quantity"] = quantity
				logEntry["byte_count"] = byteCount

				if len(data) >= int(5+byteCount) {
					logEntry["values"] = fmt.Sprintf("%02x", data[5:5+byteCount])
					m.OnWrite(startAddr, data[5:5+byteCount])
				}
			}
		}

		logJSON, _ := json.Marshal(logEntry)
		log.Printf("Modbus Request: %s", string(logJSON))

		// Call the original handler
		return originalHandler(s, frame)
	}
}

func (m *ModbusServer) Stop() error {
	log.Println("Stopping Modbus server...")
	m.server.Close()
	return nil
}
