package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/adapters/dnp3"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/mqtt"
	"github.com/fruganyumisa/igrid-middleware/internal/models"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/config"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/logger"
)

// Adapter defines the common interface for all protocol adapters
type Adapter interface {
	// Start initializes and starts the adapter
	Start(ctx context.Context) error

	// Stop gracefully shuts down the adapter
	Stop(ctx context.Context) error

	// GetProtocol returns the protocol name this adapter handles
	GetProtocol() string

	// ProcessMessage processes incoming messages
	ProcessMessage(msg models.Message) error

	// IsHealthy returns the health status of the adapter
	IsHealthy() bool

	// Name returns the name of the adapter
	Name() string

	// Handle processes a normalized message
	Handle(ctx context.Context, msg interface{}) error

	// CanHandle determines if this adapter can handle the given message
	CanHandle(msg interface{}) bool
}

// AdapterConfig holds configuration for different adapter types
type AdapterConfig struct {
	Protocol string                 `yaml:"protocol"`
	Enabled  bool                   `yaml:"enabled"`
	Config   map[string]interface{} `yaml:"config"`
}

// AdapterManager manages multiple protocol adapters
type AdapterManager struct {
	adapters map[string]Adapter
	logger   logger.Logger
	config   config.Config
}

// ModbusAdapterWrapper wraps the Modbus client to implement the Adapter interface
type ModbusAdapterWrapper struct {
	client *modbus.ModbusClient
	logger logger.Logger
}

func (m *ModbusAdapterWrapper) Start(ctx context.Context) error {
	// Start the Modbus client polling
	go m.client.PollRegisters(ctx)
	return nil
}

func (m *ModbusAdapterWrapper) Stop(ctx context.Context) error {
	// The Modbus client will stop when context is cancelled
	return nil
}

func (m *ModbusAdapterWrapper) GetProtocol() string {
	return "modbus"
}

func (m *ModbusAdapterWrapper) ProcessMessage(msg models.Message) error {
	// Process incoming Modbus messages if needed
	return nil
}

func (m *ModbusAdapterWrapper) IsHealthy() bool {
	// Check if the Modbus client is healthy
	return m.client != nil
}

func (m *ModbusAdapterWrapper) Name() string {
	return "modbus-adapter"
}

func (m *ModbusAdapterWrapper) Handle(ctx context.Context, msg interface{}) error {
	// Handle the normalized message
	m.logger.Info("Handling message in Modbus adapter", "msg", msg)
	// Add specific Modbus handling logic here
	return nil
}

func (m *ModbusAdapterWrapper) CanHandle(msg interface{}) bool {
	// Check if this message is for Modbus protocol
	if msgMap, ok := msg.(map[string]interface{}); ok {
		if protocol, exists := msgMap["protocol"]; exists {
			return protocol == "modbus"
		}
	}
	return false
}

// DNP3AdapterWrapper wraps the DNP3 handler to implement the Adapter interface
type DNP3AdapterWrapper struct {
	handler *dnp3.DNP3Handler
	logger  logger.Logger
}

func (d *DNP3AdapterWrapper) Start(ctx context.Context) error {
	return d.handler.Start(ctx)
}

func (d *DNP3AdapterWrapper) Stop(ctx context.Context) error {
	// The DNP3 handler will stop when context is cancelled
	return nil
}

func (d *DNP3AdapterWrapper) GetProtocol() string {
	return "dnp3"
}

func (d *DNP3AdapterWrapper) ProcessMessage(msg models.Message) error {
	// Process incoming DNP3 messages if needed
	return nil
}

func (d *DNP3AdapterWrapper) IsHealthy() bool {
	// Check if the DNP3 handler is healthy
	return d.handler != nil
}

func (d *DNP3AdapterWrapper) Name() string {
	return "dnp3-adapter"
}

func (d *DNP3AdapterWrapper) Handle(ctx context.Context, msg interface{}) error {
	// Handle the normalized message
	d.logger.Info("Handling message in DNP3 adapter", "msg", msg)
	// Add specific DNP3 handling logic here
	return nil
}

func (d *DNP3AdapterWrapper) CanHandle(msg interface{}) bool {
	// Check if this message is for DNP3 protocol
	if msgMap, ok := msg.(map[string]interface{}); ok {
		if protocol, exists := msgMap["protocol"]; exists {
			return protocol == "dnp3"
		}
	}
	return false
}

// MQTTAdapterWrapper wraps the MQTT publisher to implement the Adapter interface
type MQTTAdapterWrapper struct {
	publisher *mqtt.Publisher
	logger    logger.Logger
}

func (m *MQTTAdapterWrapper) Start(ctx context.Context) error {
	// MQTT publisher is already connected during creation
	return nil
}

func (m *MQTTAdapterWrapper) Stop(ctx context.Context) error {
	// The MQTT publisher will be handled by the underlying client
	// Add any cleanup logic here if needed
	return nil
}

func (m *MQTTAdapterWrapper) GetProtocol() string {
	return "mqtt"
}

func (m *MQTTAdapterWrapper) ProcessMessage(msg models.Message) error {
	// Convert message to JSON and publish via MQTT
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return m.publisher.Publish(context.Background(), jsonData)
}

func (m *MQTTAdapterWrapper) IsHealthy() bool {
	// Check if the MQTT publisher is healthy
	return m.publisher != nil
}

// Constructor functions for adapter wrappers

// NewModbusAdapterWrapper creates a new Modbus adapter wrapper
func NewModbusAdapterWrapper(client *modbus.ModbusClient, log logger.Logger) *ModbusAdapterWrapper {
	return &ModbusAdapterWrapper{
		client: client,
		logger: log,
	}
}

// NewDNP3AdapterWrapper creates a new DNP3 adapter wrapper
func NewDNP3AdapterWrapper(handler *dnp3.DNP3Handler, log logger.Logger) *DNP3AdapterWrapper {
	return &DNP3AdapterWrapper{
		handler: handler,
		logger:  log,
	}
}

// NewMQTTAdapterWrapper creates a new MQTT adapter wrapper
func NewMQTTAdapterWrapper(publisher *mqtt.Publisher, log logger.Logger) *MQTTAdapterWrapper {
	return &MQTTAdapterWrapper{
		publisher: publisher,
		logger:    log,
	}
}

func (m *MQTTAdapterWrapper) Name() string {
	return "mqtt-adapter"
}

func (m *MQTTAdapterWrapper) Handle(ctx context.Context, msg interface{}) error {
	// Handle the normalized message by publishing it via MQTT
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return m.publisher.Publish(ctx, jsonData)
}

func (m *MQTTAdapterWrapper) CanHandle(msg interface{}) bool {
	// Check if this message is for MQTT protocol or if MQTT should handle all outgoing messages
	if msgMap, ok := msg.(map[string]interface{}); ok {
		if protocol, exists := msgMap["protocol"]; exists {
			return protocol == "mqtt"
		}
		// MQTT adapter can also handle messages that need to be published regardless of protocol
		if destination, exists := msgMap["destination"]; exists {
			return destination != nil // MQTT can handle any message with a destination
		}
	}
	return true // MQTT adapter can handle most normalized messages for publishing
}

// NewAdapterManager creates a new adapter manager
func NewAdapterManager(cfg config.Config, log logger.Logger) *AdapterManager {
	return &AdapterManager{
		adapters: make(map[string]Adapter),
		logger:   log,
		config:   cfg,
	}
}

// RegisterAdapter registers an adapter with the manager
func (am *AdapterManager) RegisterAdapter(protocol string, adapter Adapter) error {
	if adapter == nil {
		return fmt.Errorf("adapter cannot be nil")
	}

	am.adapters[protocol] = adapter
	am.logger.Info("Adapter registered", "protocol", protocol)
	return nil
}

// GetAdapter retrieves an adapter by protocol name
func (am *AdapterManager) GetAdapter(protocol string) (Adapter, error) {
	adapter, exists := am.adapters[protocol]
	if !exists {
		return nil, fmt.Errorf("adapter not found for protocol: %s", protocol)
	}
	return adapter, nil
}

// StartAll starts all registered adapters
func (am *AdapterManager) StartAll(ctx context.Context) error {
	for protocol, adapter := range am.adapters {
		if err := adapter.Start(ctx); err != nil {
			am.logger.Error("Failed to start adapter", "protocol", protocol, "error", err)
			return fmt.Errorf("failed to start %s adapter: %w", protocol, err)
		}
		am.logger.Info("Adapter started successfully", "protocol", protocol)
	}
	return nil
}

// StopAll stops all registered adapters
func (am *AdapterManager) StopAll(ctx context.Context) error {
	var lastError error

	for protocol, adapter := range am.adapters {
		if err := adapter.Stop(ctx); err != nil {
			am.logger.Error("Failed to stop adapter", "protocol", protocol, "error", err)
			lastError = err
		} else {
			am.logger.Info("Adapter stopped successfully", "protocol", protocol)
		}
	}

	return lastError
}

// GetHealthStatus returns the health status of all adapters
func (am *AdapterManager) GetHealthStatus() map[string]bool {
	status := make(map[string]bool)
	for protocol, adapter := range am.adapters {
		status[protocol] = adapter.IsHealthy()
	}
	return status
}

// CreateAdapter creates an adapter based on configuration
func CreateAdapter(adapterCfg AdapterConfig, globalCfg config.Config, log logger.Logger) (Adapter, error) {
	if !adapterCfg.Enabled {
		return nil, fmt.Errorf("adapter for protocol %s is disabled", adapterCfg.Protocol)
	}

	switch adapterCfg.Protocol {
	case "modbus":
		return createModbusAdapter(adapterCfg.Config, globalCfg, log)
	case "dnp3":
		return createDNP3Adapter(adapterCfg.Config, globalCfg, log)
	case "mqtt":
		return createMQTTAdapter(adapterCfg.Config, globalCfg, log)
	default:
		return nil, fmt.Errorf("unsupported adapter protocol: %s", adapterCfg.Protocol)
	}
}

// createModbusAdapter creates a Modbus adapter with proper configuration
func createModbusAdapter(adapterCfg map[string]interface{}, globalCfg config.Config, log logger.Logger) (Adapter, error) {
	// Extract Modbus-specific configuration
	address, ok := adapterCfg["address"].(string)
	if !ok {
		return nil, fmt.Errorf("modbus adapter requires 'address' configuration")
	}

	pollInterval, ok := adapterCfg["poll_interval"].(string)
	if !ok {
		pollInterval = "5s" // default
	}

	duration, err := time.ParseDuration(pollInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid poll_interval for modbus: %w", err)
	}

	cfg := modbus.Config{
		Address:      address,
		PollInterval: duration,
		Registers:    []modbus.Register{}, // This should be populated from config
	}

	client, err := modbus.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create modbus client: %w", err)
	}

	return &ModbusAdapterWrapper{
		client: client,
		logger: log,
	}, nil
}

// createDNP3Adapter creates a DNP3 adapter with proper configuration
func createDNP3Adapter(adapterCfg map[string]interface{}, globalCfg config.Config, log logger.Logger) (Adapter, error) {
	endpoint, ok := adapterCfg["dss_endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("dnp3 adapter requires 'dss_endpoint' configuration")
	}

	cfg := dnp3.Config{
		DSSEndpoint:      endpoint,
		OutstationConfig: adapterCfg["outstation_config"],
	}

	handler := dnp3.NewHandler(cfg)

	return &DNP3AdapterWrapper{
		handler: handler,
		logger:  log,
	}, nil
}

// createMQTTAdapter creates an MQTT adapter with proper configuration
func createMQTTAdapter(adapterCfg map[string]interface{}, globalCfg config.Config, log logger.Logger) (Adapter, error) {
	publisher, err := mqtt.NewPublisher(globalCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create mqtt publisher: %w", err)
	}

	return &MQTTAdapterWrapper{
		publisher: publisher,
		logger:    log,
	}, nil
}
