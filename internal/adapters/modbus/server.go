package modbus

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/tbrandon/mbserver"
)

type ModbusServer struct {
	server *mbserver.Server
	// Add logger, MQTT/HTTP clients as needed
}

func NewServer() (*ModbusServer, error) {
	srv := mbserver.NewServer()
	return &ModbusServer{server: srv}, nil
}

func (m *ModbusServer) Start(ctx context.Context, address string) error {
	go func() {
		<-ctx.Done()
		m.server.Close()
	}()
	err := m.server.ListenTCP(address)
	if err != nil {
		return err
	}
	log.Printf("Modbus server started on %s", address)
	return nil
}

// Example: Hook for write events (customize as needed)
func (m *ModbusServer) OnWrite(address uint16, value []byte) {
	msg := Message{
		Timestamp: time.Now().UTC(),
		Source:    "modbus-server",
		Values:    map[string]interface{}{"address": address, "value": value},
	}
	jsonData, _ := json.Marshal(msg)
	log.Printf("OnWrite event: %s", jsonData)
	// TODO: Publish jsonData to MQTT/HTTP
}
