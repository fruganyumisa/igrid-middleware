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
	"github.com/fruganyumisa/igrid-middleware/internal/models"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Declare adapterList before use
	var adapterList []adapters.Adapter

	// Load configuration
	cfg, err := config.Load("configs/gateway.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	modbusServerFlag := flag.Bool("modbus-server", false, "Start Modbus TCP server")
	flag.Parse()

	// Initialize logger early
	zapLogger := logger.New(cfg.Logging.Level)

	if *modbusServerFlag {
		// Create Modbus server with smart grid data processing
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		server, err := modbus.NewServer("GATEWAY_DEVICE")
		if err != nil {
			log.Fatalf("Failed to start Modbus server: %v", err)
		}

		// Create HTTP callback with authentication and headers from configuration
		var httpCallback modbus.HTTPCallback
		if cfg.HTTP.DMS.BaseURL != "" {
			timeout := time.Duration(cfg.HTTP.DMS.Timeout) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second // default timeout
			}

			// Extract auth configuration
			authType := cfg.HTTP.DMS.AuthType
			authToken := cfg.HTTP.DMS.AuthToken
			headers := cfg.HTTP.DMS.Headers

			// Log DMS configuration
			fmt.Printf("🔧 Configuring DMS integration:\n")
			fmt.Printf("   Base URL: %s\n", cfg.HTTP.DMS.BaseURL)
			fmt.Printf("   Endpoint: %s\n", cfg.HTTP.DMS.Endpoint)
			fmt.Printf("   Auth Type: %s\n", authType)
			fmt.Printf("   Timeout: %v\n", timeout)
			fmt.Printf("   Custom Headers: %v\n", headers)

			// Create advanced HTTP callback with configuration
			httpCallback = modbus.CreateAdvancedHTTPCallback(
				cfg.HTTP.DMS.BaseURL,
				timeout,
				authType,
				authToken,
				headers,
			)
			fmt.Println("✅ DMS integration configured successfully")
		} else {
			fmt.Println("⚠️  No DMS configuration found - HTTP integration disabled")
		}

		// Set the HTTP callback for DMS integration
		if httpCallback != nil {
			server.SetHTTPCallback(httpCallback)
		}

		// Optional: Set data callback for additional processing/logging
		server.SetDataCallback(func(data *models.SmartGridData) error {
			fmt.Printf("📊 Smart Grid Data: Device=%s, V=%.2fV, I=%.3fA, T=%.2f°C\n",
				data.DeviceID, data.Voltage, data.Current, data.Temperature)
			return nil
		})

		go func() {
			if err := server.Start(ctx, ":502"); err != nil {
				fmt.Printf("Modbus server stopped: %v\n", err)
			}
		}()
		fmt.Println("🚀 Modbus server started on :502 with smart grid data processing")

		// Wait for shutdown signal when running as server only
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Shutting down...")
		return
	}

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
			zapLogger.Error("Failed to create normalizer even without schema", zap.Error(err))
			return
		}
	}
	zapLogger.Info("Normalizer initialized", zap.Any("normalizer", normalizer))

	// Create Modbus adapter if enabled
	if cfg.Modbus.Enabled {
		// Create Modbus client with smart grid register configuration
		modbusConfig := modbus.Config{
			Address: fmt.Sprintf("%s:%d", cfg.Modbus.Host, cfg.Modbus.Port),
			Registers: []modbus.Register{
				// Power Meter Registers
				{
					Name:     "voltage_phase_a",
					Address:  40001,
					Quantity: 2,
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
					Address:  40003,
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
					Address:  40005,
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
					Address:  40007,
					Quantity: 1,
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
					Address:  40009,
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
			zapLogger.Error("Failed to create Modbus client", zap.Error(err))
		} else {
			modbusAdapter := adapters.NewModbusAdapterWrapper(modbusClient, zapLogger)
			adapterList = append(adapterList, modbusAdapter)
			zapLogger.Info("Modbus adapter created")
		}
	}

	// Create DNP3 adapter if enabled
	if cfg.DNP3.Enabled {
		// Only use fields that exist in dnp3.Config struct
		dnp3Config := dnp3.Config{
			// Replace these with actual fields from your dnp3.Config definition
			// For example, if your struct has Address and Timeout:
			// Address: cfg.DNP3.Address,
			// Timeout: cfg.DNP3.Timeout,
		}
		dnp3Handler := dnp3.NewHandler(dnp3Config)
		dnp3Adapter := adapters.NewDNP3AdapterWrapper(dnp3Handler, zapLogger)
		adapterList = append(adapterList, dnp3Adapter)
		zapLogger.Info("DNP3 adapter created")
	}

	// Skip MQTT adapter for now due to config complexity
	// TODO: Fix config type mismatch between root config and internal config
	if cfg.MQTT.Enabled {
		fmt.Println("MQTT adapter creation skipped due to config type mismatch - TODO: Fix")
	}

	// Create router with adapters
	router := core.NewRouter(adapterList, zapLogger)
	if router == nil {
		zapLogger.Error("Failed to create router")
		return
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all adapters
	for _, adapter := range adapterList {
		go func(a adapters.Adapter) {
			err := a.Start(ctx)
			if err != nil {
				zapLogger.Error("Failed to start adapter", zap.String("adapter", a.Name()), zap.Error(err))
				cancel()
			} else {
				zapLogger.Info("Started adapter", zap.String("name", a.Name()))
			}
		}(adapter)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	zapLogger.Info("Shutting down...")

	// Stop all adapters
	for _, adapter := range adapterList {
		err := adapter.Stop(ctx)
		if err != nil {
			zapLogger.Error("Failed to stop adapter", zap.String("adapter", adapter.Name()), zap.Error(err))
		}
	}
}
