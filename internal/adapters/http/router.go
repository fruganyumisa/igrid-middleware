package http

import (
	"context"
	"log"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/models"
)

type SmartGridData struct {
	DeviceID      string    `json:"deviceId"`
	Voltage       float64   `json:"voltage"`
	Current       float64   `json:"current"`
	ActivePower   float64   `json:"active_power"`
	ReactivePower float64   `json:"reactive_power"`
	Frequency     float64   `json:"frequency"`
	BreakerStatus uint16    `json:"breaker_status"`
	TapPosition   uint16    `json:"tap_position"`
	Energy        float64   `json:"energy"`
	Temperature   float64   `json:"temperature"`
	Humidity      float64   `json:"humidity"`
	Alarm         uint16    `json:"alarm"`
	Timestamp     int64     `json:"timestamp"`
	ModbusAddress uint16    `json:"modbus_address"`
	ReceivedAt    time.Time `json:"received_at"`
}

// HTTPRouter handles routing smart grid data to HTTP endpoints
type HTTPRouter struct {
	httpClient *HTTPClient
}

// NewHTTPRouter creates a new HTTP router with the given HTTP client
func NewHTTPRouter(httpClient *HTTPClient) *HTTPRouter {
	return &HTTPRouter{
		httpClient: httpClient,
	}
}

// RouteSmartGridData implements the DataRouter interface
func (r *HTTPRouter) RouteSmartGridData(ctx context.Context, data *SmartGridData) error {
	// Convert to DMS payload format
	dmsPayload := &models.DMSPayload{
		DeviceID:    data.DeviceID,
		Voltage:     data.Voltage,
		Current:     data.Current,
		Temperature: data.Temperature,
	}

	if err := r.httpClient.SendToDMS(ctx, dmsPayload); err != nil {
		log.Printf("Failed to send data to DMS: %v", err)
		return err
	}

	log.Printf("Successfully routed smart grid data for device %s to DMS", data.DeviceID)
	return nil
}

// toDMSPayload converts SmartGridData to DMS application format
func (data *SmartGridData) toDMSPayload() *models.DMSPayload {
	return &models.DMSPayload{
		DeviceID:    data.DeviceID,
		Voltage:     data.Voltage,
		Current:     data.Current,
		Temperature: data.Temperature,
	}
}
