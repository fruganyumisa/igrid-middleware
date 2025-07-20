package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fruganyumisa/igrid-middleware//internal/adapters/modbus"
	"github.com/fruganyumisa/igrid-middleware//internal/pkg/logger"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/dnp3"
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/mqtt"
	"github.com/fruganyumisa/igrid-middleware/internal/core"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/config"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/gateway.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Create normalizer with protocol mappings
	normalizer, err := core.NewNormalizer(cfg.Mappings, cfg.SchemaPath)
	if err != nil {
		zapLogger.Fatal("Failed to create normalizer", zap.Error(err))
	}

	// Create MQTT publisher
	mqttPublisher, err := mqtt.NewPublisher(cfg.MQTT)
	if err != nil {
		zapLogger.Fatal("Failed to create MQTT publisher", zap.Error(err))
	}

	// Initialize protocol adapters
	dnp3Handler := dnp3.NewHandler(cfg.DNP3, normalizer, zapLogger)
	modbusClient := modbus.NewClient(cfg.Modbus, zapLogger)

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

	go func() {
		modbusClient.PollRegisters(ctx)
	}()

	// HTTP server
	httpServer := NewHTTPServer(cfg.HTTP, normalizer, zapLogger)
	go func() {
		if err := httpServer.Start(); err != nil {
			zapLogger.Error("HTTP server failed", zap.Error(err))
			cancel()
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	zapLogger.Info("Shutting down...")
}
