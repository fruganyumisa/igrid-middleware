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

func (m *ModbusServer) logRequest(funcCode uint8, params map[string]interface{}) {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	logEntry := map[string]interface{}{
		"timestamp":      timestamp,
		"protocol":       "modbus",
		"function_code":  funcCode,
		"function_name":  getFunctionName(funcCode),
		"parameters":     params,
		"server_address": m.addr,
	}

	logJSON, _ := json.Marshal(logEntry)
	log.Printf("Modbus Request: %s", string(logJSON))
}

func (m *ModbusServer) Start(ctx context.Context, address string) error {
	m.addr = address

	log.Printf("Starting Modbus TCP server on %s with comprehensive request logging", address)

	go func() {
		<-ctx.Done()
		log.Println("Shutting down Modbus server...")
		m.server.Close()
	}()

	return m.server.ListenTCP(address)
}

func (m *ModbusServer) Stop() error {
	log.Println("Stopping Modbus server...")
	m.server.Close()
	return nil
}
