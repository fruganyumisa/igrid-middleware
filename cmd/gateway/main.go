package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fruganyumisa/igrid-middleware/config"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/dnp3"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware/internal/core"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/gateway.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	modbusServerFlag := flag.Bool("modbus-server", false, "Start Modbus TCP server")
	flag.Parse()

	if *modbusServerFlag {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		server, err := modbus.NewServer()
		if err != nil {
			log.Fatalf("Failed to start Modbus server: %v", err)
		}
		go func() {
			if err := server.Start(ctx, ":502"); err != nil {
				log.Printf("Modbus server stopped: %v", err)
			}
		}()
		log.Println("Modbus server started on :502")
	}

	// Initialize logger
	zapLogger := logger.New(cfg.Logging.Level)

	// Create normalizer with protocol mappings
	normalization := make(map[string]core.MappingConfig)
	for k, v := range cfg.Normalization {
		zapLogger.Info("Adding normalization mapping", zap.String("source", v.SourceField), zap.String("target", v.TargetField))
		normalization[k] = core.MappingConfig{
			JSONPath: v.JSONPath,
			Scaling:  v.Scaling,
			DataType: v.DataType,
			Unit:     v.Unit,
		}
	}
	normalizer, err := core.NewNormalizer(normalization, cfg.SchemaPath)
	if err != nil {
		zapLogger.Error("Failed to create normalizer", zap.Error(err))
		zapLogger.Info("Continuing without schema validation")
		// Create normalizer without schema validation for now
		normalizer, err = core.NewNormalizer(normalization, "")
		if err != nil {
			zapLogger.Fatal("Failed to create normalizer even without schema", zap.Error(err))
		}
	}
	zapLogger.Info("Normalizer initialized", zap.Any("normalizer", normalizer))

	// Create adapter manager
	// For now, create adapters directly without the manager due to config type mismatch
	var adapterList []adapters.Adapter

	// Create Modbus adapter if enabled
	if cfg.Modbus.Enabled {
		// Create Modbus client with smart grid register configuration
		modbusConfig := modbus.Config{
			Address:      fmt.Sprintf("%s:%d", cfg.Modbus.Host, cfg.Modbus.Port),
			PollInterval: 5 * time.Second,
			Registers: []modbus.Register{
				// Power Meter Registers
				{
					Name:     "voltage_phase_a",
					Address:  40001, // Holding register for Phase A Voltage
					Quantity: 2,     // 2 registers for 32-bit float
					TransformValue: func(data []byte) interface{} {
						// Convert 2 registers to float32 (voltage in volts)
						if len(data) >= 4 {
							// Assuming big-endian format
							value := float32(uint32(data[0])<<24|uint32(data[1])<<16|uint32(data[2])<<8|uint32(data[3])) / 100.0
							return value
						}
						return 0.0
					},
				},
				{
					Name:     "current_phase_a",
					Address:  40003, // Current measurement
					Quantity: 2,
					TransformValue: func(data []byte) interface{} {
						// Convert to amperes
						if len(data) >= 4 {
							value := float32(uint32(data[0])<<24|uint32(data[1])<<16|uint32(data[2])<<8|uint32(data[3])) / 1000.0
							return value
						}
						return 0.0
					},
				},
				{
					Name:     "active_power",
					Address:  40005, // Active power in watts
					Quantity: 2,
					TransformValue: func(data []byte) interface{} {
						if len(data) >= 4 {
							value := float32(uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]))
							return value
						}
						return 0.0
					},
				},
				{
					Name:     "frequency",
					Address:  40007, // Grid frequency
					Quantity: 1,     // Single register
					TransformValue: func(data []byte) interface{} {
						if len(data) >= 2 {
							value := float32(uint16(data[0])<<8|uint16(data[1])) / 100.0 // Frequency in Hz
							return value
						}
						return 50.0 // Default frequency
					},
				},
				// Energy Meter Registers
				{
					Name:     "total_energy",
					Address:  40009, // Total energy consumed
					Quantity: 2,
					TransformValue: func(data []byte) interface{} {
						if len(data) >= 4 {
							value := float32(uint32(data[0])<<24|uint32(data[1])<<16|uint32(data[2])<<8|uint32(data[3])) / 1000.0 // kWh
							return value
						}
						return 0.0
					},
				},
				// Temperature Sensor
				{
					Name:     "transformer_temp",
					Address:  40011,
					Quantity: 1,
					TransformValue: func(data []byte) interface{} {
						if len(data) >= 2 {
							value := float32(int16(uint16(data[0])<<8|uint16(data[1]))) / 10.0 // Temperature in Celsius
							return value
						}
						return 0.0
					},
				},
				// Status Registers
				{
					Name:     "device_status",
					Address:  40013,
					Quantity: 1,
					TransformValue: func(data []byte) interface{} {
						if len(data) >= 2 {
							status := uint16(data[0])<<8 | uint16(data[1])
							statusMap := map[string]bool{
								"online":      (status & 0x0001) != 0,
								"fault":       (status & 0x0002) != 0,
								"maintenance": (status & 0x0004) != 0,
								"alarm":       (status & 0x0008) != 0,
							}
							return statusMap
						}
						return map[string]bool{"online": false}
					},
				},
			},
		}
		modbusClient, err := modbus.NewClient(modbusConfig)
		if err != nil {
			zapLogger.Fatal("Failed to create Modbus client", zap.Error(err))
		}
		modbusAdapter := adapters.NewModbusAdapterWrapper(modbusClient, zapLogger)
		adapterList = append(adapterList, modbusAdapter)
		zapLogger.Info("Modbus adapter created")
	}

	// Create DNP3 adapter if enabled
	if cfg.DNP3.Enabled {
		dnp3Config := dnp3.Config{
			DSSEndpoint: fmt.Sprintf("http://%s:%d", cfg.DNP3.Host, cfg.DNP3.Port),
		}
		dnp3Handler := dnp3.NewHandler(dnp3Config)
		dnp3Adapter := adapters.NewDNP3AdapterWrapper(dnp3Handler, zapLogger)
		adapterList = append(adapterList, dnp3Adapter)
		zapLogger.Info("DNP3 adapter created")
	}

	// Skip MQTT adapter for now due to config complexity
	// TODO: Fix config type mismatch between root config and internal config
	if cfg.MQTT.Enabled {
		zapLogger.Info("MQTT adapter creation skipped due to config type mismatch - TODO: Fix")
	}

	// Create router with adapters
	router := core.NewRouter(adapterList, zapLogger)
	if router == nil {
		zapLogger.Fatal("Failed to create router")
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all adapters
	for _, adapter := range adapterList {
		go func(a adapters.Adapter) {
			if err := a.Start(ctx); err != nil {
				zapLogger.Error("Failed to start adapter", zap.String("adapter", a.Name()), zap.Error(err))
				cancel()
			}
		}(adapter)
		zapLogger.Info("Started adapter", zap.String("name", adapter.Name()))
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	zapLogger.Info("Shutting down...")

	// Stop all adapters
	for _, adapter := range adapterList {
		if err := adapter.Stop(ctx); err != nil {
			zapLogger.Error("Failed to stop adapter", zap.String("adapter", adapter.Name()), zap.Error(err))
		}
	}

}
