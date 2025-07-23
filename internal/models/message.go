package models

import "time"

type Message struct {
	ID          string                 `json:"id" validate:"required,uuid4"`
	Timestamp   time.Time              `json:"timestamp" validate:"required"`
	Source      string                 `json:"source" validate:"required"`
	Destination string                 `json:"destination" validate:"required"`
	Protocol    string                 `json:"protocol" validate:"required,oneof=modbus dnp3 mqtt http"`
	Payload     map[string]interface{} `json:"payload" validate:"required"`
	Metadata    map[string]string      `json:"metadata"`
}

type MessageBatch struct {
	Messages []Message `json:"messages" validate:"required,dive"`
}

// SmartGridData represents the normalized smart grid data structure
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

// DMSPayload represents the payload structure for DMS application
type DMSPayload struct {
	DeviceID    string  `json:"deviceId"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Temperature float64 `json:"temperature"`
}
