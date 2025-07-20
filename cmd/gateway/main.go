package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fruganyumisa/igrid-middleware/internal/adapters/dnp3"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware/internal/core"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/config"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/gateway.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dnp3Config := dnp3.Config{
		// Replace 'DSSEndpoint' with the correct field name from config.DNP3Config, for example:
		// Endpoint: cfg.DNP3.Endpoint,
		// Copy other fields as needed
		DSSEndpoint: "NULL", // Placeholder, replace with actual config field
		// OutstationConfig: cfg.DNP3.OutstationConfig, // Adjust as per
	}

	dnp3Handler := dnp3.NewHandler(dnp3Config)

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
		zapLogger.Info("Adding normalization mapping", zap.String("source", v.Source), zap.String("destination", v.Destination))
		normalization[k] = core.MappingConfig{
			JSONPath: v.Source,
		}
	}
	normalizer, err := core.NewNormalizer(normalization, cfg.SchemaPath)
	zapLogger.Info("Normalizer initialized", zap.Any("normalizer", normalizer))
	if err != nil {
		zapLogger.Fatal("Failed to create normalizer", zap.Error(err))
	}

	// start the Modbus server
	modbusServer, err := modbus.NewServer()
	if err != nil {
		zapLogger.Fatal("Failed to create Modbus server", zap.Error(err))
	}
	if err := modbusServer.Start(context.Background(), cfg.Modbus.Address); err != nil {
		zapLogger.Fatal("Failed to start Modbus server", zap.Error(err))
	}

	// Create MQTT publisher
	// mqttPublisher, err := mqtt.NewPublisher(cfg.MQTT)
	// if err != nil {
	// 	zapLogger.Fatal("Failed to create MQTT publisher", zap.Error(err))
	// }

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start adapters
	go func() {
		if err := dnp3Handler.Start(ctx); err != nil {
			zapLogger.Error("DNP3 handler failed", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	zapLogger.Info("Shutting down...")

}
