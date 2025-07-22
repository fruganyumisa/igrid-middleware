package modbus

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goburrow/modbus"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Config struct {
	Address      string
	PollInterval time.Duration
	Registers    []Register
}

type Register struct {
	Name           string
	Address        uint16
	Quantity       uint16
	TransformValue func([]byte) interface{}
}

type ModbusClient struct {
	client   modbus.Client
	logger   *zap.Logger
	mqttChan *amqp.Channel
	config   Config
}

func NewClient(cfg Config) (*ModbusClient, error) {
	handler := modbus.NewTCPClientHandler(cfg.Address)
	handler.Timeout = 10 * time.Second

	if err := handler.Connect(); err != nil {
		return nil, err
	}

	return &ModbusClient{
		client: modbus.NewClient(handler),
		config: cfg,
	}, nil
}

func (m *ModbusClient) Publish(data []byte) {
	if m.mqttChan == nil {
		if m.logger != nil {
			m.logger.Error("AMQP channel not initialized")
		}
		return
	}
	err := m.mqttChan.Publish(
		"",       // exchange
		"modbus", // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil && m.logger != nil {
		m.logger.Error("Failed to publish message", zap.Error(err))
	}
}

// Message struct removed to avoid redeclaration error.
// Import from server.go if needed.

func (m *ModbusClient) PollRegisters(ctx context.Context) {
	ticker := time.NewTicker(m.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results := make(map[string]interface{})
			for _, reg := range m.config.Registers {
				val, err := m.client.ReadHoldingRegisters(reg.Address, reg.Quantity)
				if err != nil {
					if m.logger != nil {
						m.logger.Error("Modbus read failed", zap.Error(err))
					}
					continue
				}
				// Apply scaling and type conversion
				results[reg.Name] = reg.TransformValue(val)
			}

			msg := Message{
				Timestamp: time.Now().UTC(),
				Source:    "modbus",
				Values:    results,
			}

			jsonData, _ := json.Marshal(msg)
			m.Publish(jsonData)
		}
	}
}
