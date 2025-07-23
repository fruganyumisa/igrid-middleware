package modbus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/models"
	"github.com/tbrandon/mbserver"
)

// DataCallback function type for handling processed smart grid data
type DataCallback func(data *models.SmartGridData) error

// HTTPCallback function type for sending data to HTTP endpoints
type HTTPCallback func(endpoint string, payload *models.DMSPayload) error

type ModbusServer struct {
	server        *mbserver.Server
	addr          string
	dataCallback  DataCallback      // Optional callback for processed data
	httpCallback  HTTPCallback      // Optional callback for HTTP routing
	registerCache map[uint16]uint16 // Cache for register values
	deviceID      string            // Device identifier
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

func NewServer(deviceID string) (*ModbusServer, error) {
	srv := mbserver.NewServer()

	modbusServer := &ModbusServer{
		server:        srv,
		addr:          ":5020",
		registerCache: make(map[uint16]uint16),
		deviceID:      deviceID,
	}

	// Initialize with sample smart grid data
	srv.HoldingRegisters[0] = 26247  // voltage: 262.47V
	srv.HoldingRegisters[1] = 14560  // current: 14.56A
	srv.HoldingRegisters[2] = 15210  // active_power: 1521.0W
	srv.HoldingRegisters[3] = 12280  // reactive_power: 122.80VAR
	srv.HoldingRegisters[4] = 5034   // frequency: 50.34Hz
	srv.HoldingRegisters[5] = 0      // breaker_status: open
	srv.HoldingRegisters[6] = 22     // tap_position: 22
	srv.HoldingRegisters[7] = 1443   // energy high word (23274.49 split)
	srv.HoldingRegisters[8] = 61711  // energy low word
	srv.HoldingRegisters[9] = 6370   // temperature: 63.70°C
	srv.HoldingRegisters[10] = 6048  // humidity: 60.48%
	srv.HoldingRegisters[11] = 0     // alarm: no alarm
	srv.HoldingRegisters[12] = 26696 // timestamp high word (1753227100 split)
	srv.HoldingRegisters[13] = 8300  // timestamp low word
	srv.HoldingRegisters[14] = 0     // device register (will be set from deviceID)

	// Set input registers with same data for testing
	for i := 0; i < 15; i++ {
		srv.InputRegisters[i] = srv.HoldingRegisters[i]
	}

	srv.Debug = true

	return modbusServer, nil
}

// SetDataCallback sets a callback function for handling processed smart grid data
func (m *ModbusServer) SetDataCallback(callback DataCallback) {
	m.dataCallback = callback
}

// SetHTTPCallback sets a callback function for HTTP routing
func (m *ModbusServer) SetHTTPCallback(callback HTTPCallback) {
	m.httpCallback = callback
}

// SendToDMS sends data to DMS endpoint using HTTP callback
func (m *ModbusServer) SendToDMS(payload *models.DMSPayload) error {
	if m.httpCallback != nil {
		return m.httpCallback("/api/v1/devices/data", payload)
	}
	log.Println("No HTTP callback configured - would send to DMS:", payload)
	return nil
}

// SendHTTPRequest sends HTTP POST request to specified endpoint with JSON payload
func SendHTTPRequest(endpoint string, payload interface{}, timeout time.Duration) error {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: timeout}

	// Create request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	log.Printf("🌐 Sending HTTP POST to %s: %s", endpoint, string(jsonData))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("📨 HTTP Response [%d]: %s", resp.StatusCode, string(respBody))

	// Check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CreateHTTPCallback creates an HTTP callback function for a specific base URL
func CreateHTTPCallback(baseURL string, timeout time.Duration) HTTPCallback {
	return func(endpoint string, payload *models.DMSPayload) error {
		fullURL := baseURL + endpoint
		return SendHTTPRequest(fullURL, payload, timeout)
	}
}

// CreateAdvancedHTTPCallback creates an HTTP callback with authentication and custom headers
func CreateAdvancedHTTPCallback(baseURL string, timeout time.Duration, authType, authToken string, headers map[string]string) HTTPCallback {
	return func(endpoint string, payload *models.DMSPayload) error {
		fullURL := baseURL + endpoint
		return SendAdvancedHTTPRequest(fullURL, payload, timeout, authType, authToken, headers)
	}
}

// SendAdvancedHTTPRequest sends HTTP POST request with authentication and custom headers
func SendAdvancedHTTPRequest(url string, payload interface{}, timeout time.Duration, authType, authToken string, customHeaders map[string]string) error {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Create HTTP client with timeout
	client := &http.Client{Timeout: timeout}
	
	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "iGrid-Middleware/1.0")
	
	// Set custom headers
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}
	
	// Set authentication
	switch authType {
	case "bearer":
		if authToken != "" {
			req.Header.Set("Authorization", "Bearer "+authToken)
		}
	case "apikey":
		if authToken != "" {
			req.Header.Set("X-API-Key", authToken)
		}
	case "basic":
		// For basic auth, authToken should be "username:password"
		if authToken != "" {
			req.Header.Set("Authorization", "Basic "+authToken)
		}
	case "none":
		// No authentication
	default:
		log.Printf("⚠️  Unknown auth type: %s", authType)
	}
	
	log.Printf("🌐 Sending HTTP POST to %s with auth type: %s", url, authType)
	log.Printf("📤 Payload: %s", string(jsonData))
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	log.Printf("📨 HTTP Response [%d]: %s", resp.StatusCode, string(respBody))
	
	// Check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return nil
}

// ProcessSmartGridData processes smart grid data and routes it appropriately
func (m *ModbusServer) ProcessSmartGridData(data *models.SmartGridData) error {
	// Log the extracted data
	dataJSON, _ := json.Marshal(data)
	log.Printf("📊 Processed Smart Grid Data: %s", string(dataJSON))

	// Convert to DMS payload format
	dmsPayload := &models.DMSPayload{
		DeviceID:    data.DeviceID,
		Voltage:     data.Voltage,
		Current:     data.Current,
		Temperature: data.Temperature,
	}

	// Log DMS payload
	payloadJSON, _ := json.Marshal(dmsPayload)
	log.Printf("🚀 DMS Payload: %s", string(payloadJSON))

	// Call data callback if available
	if m.dataCallback != nil {
		if err := m.dataCallback(data); err != nil {
			log.Printf("Data callback error: %v", err)
		}
	}

	// Send to DMS
	if err := m.SendToDMS(dmsPayload); err != nil {
		log.Printf("Failed to send to DMS: %v", err)
		return err
	}

	log.Printf("✅ Successfully processed smart grid data for device: %s", data.DeviceID)
	return nil
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

// interpretRegisterValue provides human-readable interpretations of register values
// based on your specific smart grid data format
func (m *ModbusServer) interpretRegisterValue(address uint16, value uint16) map[string]interface{} {
	interpretations := make(map[string]interface{})

	// Smart grid specific register interpretations based on your data format
	switch address {
	case 0: // Voltage register
		voltage := float64(value) / 100.0 // Scale for voltage (e.g., 26247 → 262.47V)
		interpretations["voltage"] = voltage
		interpretations["unit"] = "V"
		interpretations["description"] = "Line voltage measurement"

	case 1: // Current register
		current := float64(value) / 1000.0 // Scale for current (e.g., 14560 → 14.56A)
		interpretations["current"] = current
		interpretations["unit"] = "A"
		interpretations["description"] = "Line current measurement"

	case 2: // Active power register
		activePower := float64(value) / 10.0 // Scale for power (e.g., 15210 → 1521.0W)
		interpretations["active_power"] = activePower
		interpretations["unit"] = "W"
		interpretations["description"] = "Active power consumption"

	case 3: // Reactive power register
		reactivePower := float64(value) / 100.0 // Scale for reactive power
		interpretations["reactive_power"] = reactivePower
		interpretations["unit"] = "VAR"
		interpretations["description"] = "Reactive power"

	case 4: // Frequency register
		frequency := float64(value) / 100.0 // Scale for frequency (e.g., 5034 → 50.34Hz)
		interpretations["frequency"] = frequency
		interpretations["unit"] = "Hz"
		interpretations["description"] = "Grid frequency"

	case 5: // Breaker status register
		interpretations["breaker_status"] = value
		interpretations["breaker_open"] = value == 0
		interpretations["breaker_closed"] = value == 1
		interpretations["description"] = "Circuit breaker status"

	case 6: // Tap position register
		interpretations["tap_position"] = value
		interpretations["description"] = "Transformer tap position"

	case 7: // Energy register (high word)
	case 8: // Energy register (low word) - combine for 32-bit energy value
		interpretations["energy_raw"] = value
		interpretations["description"] = "Energy measurement register"

	case 9: // Temperature register
		temperature := float64(value) / 100.0 // Scale for temperature (e.g., 6370 → 63.70°C)
		interpretations["temperature"] = temperature
		interpretations["unit"] = "°C"
		interpretations["description"] = "Equipment temperature"

	case 10: // Humidity register
		humidity := float64(value) / 100.0 // Scale for humidity
		interpretations["humidity"] = humidity
		interpretations["unit"] = "%"
		interpretations["description"] = "Relative humidity"

	case 11: // Alarm register
		interpretations["alarm"] = value
		interpretations["alarm_active"] = value != 0
		interpretations["description"] = "Alarm status register"

	case 12: // Timestamp register (high word)
	case 13: // Timestamp register (low word)
		interpretations["timestamp_raw"] = value
		interpretations["description"] = "Timestamp register"

	case 14: // Device ID register
		interpretations["device_raw"] = value
		interpretations["description"] = "Device identifier"

	default:
		// Generic interpretations for any register
		interpretations["scaled_0_100"] = float64(value) / 655.35 // 0-65535 as 0-100%
		interpretations["signed_value"] = int16(value)            // Interpret as signed
		interpretations["bcd_value"] = m.decodeBCD(value)         // Try BCD decoding
	}

	// Add common bit field interpretations
	interpretations["status_bits"] = m.decodeStatusBits(value)

	// Add timestamp for when this interpretation was made
	interpretations["interpreted_at"] = time.Now().UTC().Format(time.RFC3339Nano)

	return interpretations
}

// decodeBCD attempts to decode a value as Binary Coded Decimal
func (m *ModbusServer) decodeBCD(value uint16) interface{} {
	// BCD decoding: each nibble represents a decimal digit
	thousands := (value >> 12) & 0xF
	hundreds := (value >> 8) & 0xF
	tens := (value >> 4) & 0xF
	ones := value & 0xF

	// Check if all nibbles are valid BCD (0-9)
	if thousands <= 9 && hundreds <= 9 && tens <= 9 && ones <= 9 {
		return map[string]interface{}{
			"valid":         true,
			"decimal_value": thousands*1000 + hundreds*100 + tens*10 + ones,
			"digits":        []uint16{thousands, hundreds, tens, ones},
		}
	}

	return map[string]interface{}{
		"valid":  false,
		"reason": "contains invalid BCD digits",
	}
}

// decodeStatusBits interprets the value as a collection of status bits
func (m *ModbusServer) decodeStatusBits(value uint16) map[string]interface{} {
	bits := make(map[string]interface{})

	// Common status bit interpretations in industrial automation
	statusNames := []string{
		"system_ready", "alarm_active", "fault_condition", "manual_mode",
		"auto_mode", "maintenance_required", "overload", "emergency_stop",
		"door_open", "safety_ok", "motor_running", "heater_on",
		"pump_active", "valve_open", "pressure_ok", "temperature_ok",
	}

	activeBits := make([]string, 0)
	bitDetails := make(map[string]interface{})

	for i := 0; i < 16; i++ {
		isSet := (value>>i)&1 == 1
		bitName := fmt.Sprintf("bit_%d", i)

		if i < len(statusNames) {
			bitName = statusNames[i]
		}

		bitDetails[bitName] = isSet

		if isSet {
			activeBits = append(activeBits, bitName)
		}
	}

	bits["active_bits"] = activeBits
	bits["bit_details"] = bitDetails
	bits["active_count"] = len(activeBits)

	return bits
}

// extractSmartGridData extracts and normalizes data from multiple register reads
func (m *ModbusServer) extractSmartGridData(registers map[uint16]uint16, deviceID string, modbusAddr uint16) *models.SmartGridData {
	data := &models.SmartGridData{
		DeviceID:      deviceID,
		ModbusAddress: modbusAddr,
		ReceivedAt:    time.Now(),
	}

	// Extract and scale values based on register addresses
	if val, exists := registers[0]; exists {
		data.Voltage = float64(val) / 100.0 // Scale: 26247 → 262.47V
	}

	if val, exists := registers[1]; exists {
		data.Current = float64(val) / 1000.0 // Scale: 14560 → 14.56A
	}

	if val, exists := registers[2]; exists {
		data.ActivePower = float64(val) / 10.0 // Scale: 15210 → 1521.0W
	}

	if val, exists := registers[3]; exists {
		data.ReactivePower = float64(val) / 100.0 // Scale reactive power
	}

	if val, exists := registers[4]; exists {
		data.Frequency = float64(val) / 100.0 // Scale: 5034 → 50.34Hz
	}

	if val, exists := registers[5]; exists {
		data.BreakerStatus = val
	}

	if val, exists := registers[6]; exists {
		data.TapPosition = val
	}

	// Energy might be stored as 32-bit value across two registers
	if highVal, existsHigh := registers[7]; existsHigh {
		if lowVal, existsLow := registers[8]; existsLow {
			energy32 := uint32(highVal)<<16 | uint32(lowVal)
			data.Energy = float64(energy32) / 100.0 // Scale energy value
		}
	}

	if val, exists := registers[9]; exists {
		data.Temperature = float64(val) / 100.0 // Scale: 6370 → 63.70°C
	}

	if val, exists := registers[10]; exists {
		data.Humidity = float64(val) / 100.0 // Scale humidity
	}

	if val, exists := registers[11]; exists {
		data.Alarm = val
	}

	// Timestamp might be stored as 32-bit value across two registers
	if highVal, existsHigh := registers[12]; existsHigh {
		if lowVal, existsLow := registers[13]; existsLow {
			timestamp32 := uint32(highVal)<<16 | uint32(lowVal)
			data.Timestamp = int64(timestamp32)
		}
	}

	return data
}

// processAndRouteData processes a complete set of register reads and routes to HTTP adapter
func (m *ModbusServer) processAndRouteData(ctx context.Context, registers map[uint16]uint16) {
	// Extract smart grid data from registers
	smartGridData := m.extractSmartGridData(registers, m.deviceID, 0)

	// Process and route the data using internal methods
	if err := m.ProcessSmartGridData(smartGridData); err != nil {
		log.Printf("Failed to process smart grid data: %v", err)
	}
}

// updateRegisterCache updates the register cache and checks if we have a complete set
func (m *ModbusServer) updateRegisterCache(startAddr uint16, values []uint16) {
	// Update cache with new values
	for i, value := range values {
		m.registerCache[startAddr+uint16(i)] = value
	}

	// Check if we have a complete smart grid dataset (registers 0-14)
	completeSet := make(map[uint16]uint16)
	hasCompleteSet := true

	for addr := uint16(0); addr <= 14; addr++ {
		if value, exists := m.registerCache[addr]; exists {
			completeSet[addr] = value
		} else {
			hasCompleteSet = false
			break
		}
	}

	// If we have complete set, process and route it
	if hasCompleteSet {
		log.Printf("Complete smart grid dataset detected, processing...")
		go m.processAndRouteData(context.Background(), completeSet)
	}
}

func (m *ModbusServer) Start(ctx context.Context, address string) error {
	m.addr = address

	log.Printf("Starting Modbus TCP server on %s with comprehensive request logging", address)

	// Enable debug mode for more verbose logging
	m.server.Debug = true

	// Register function handlers that will log requests
	m.server.RegisterFunctionHandler(1, m.logAndHandle(1, mbserver.ReadCoils))
	m.server.RegisterFunctionHandler(2, m.logAndHandle(2, mbserver.ReadDiscreteInputs))
	m.server.RegisterFunctionHandler(3, m.logAndHandle(3, mbserver.ReadHoldingRegisters))
	m.server.RegisterFunctionHandler(4, m.logAndHandle(4, mbserver.ReadInputRegisters))
	m.server.RegisterFunctionHandler(5, m.logAndHandle(5, mbserver.WriteSingleCoil))
	m.server.RegisterFunctionHandler(6, m.logAndHandle(6, mbserver.WriteHoldingRegister))
	m.server.RegisterFunctionHandler(15, m.logAndHandle(15, mbserver.WriteMultipleCoils))
	m.server.RegisterFunctionHandler(16, m.logAndHandle(16, mbserver.WriteHoldingRegisters))

	go func() {
		<-ctx.Done()
		log.Println("Shutting down Modbus server...")
		m.server.Close()
	}()

	// Start server and log the fact that it's listening
	log.Printf("Modbus server listening on %s - all requests will be logged", address)
	return m.server.ListenTCP(address)
}

// logAndHandle creates a wrapper function that logs requests before calling the original handler
func (m *ModbusServer) logAndHandle(funcCode uint8, originalHandler func(*mbserver.Server, mbserver.Framer) ([]byte, *mbserver.Exception)) func(*mbserver.Server, mbserver.Framer) ([]byte, *mbserver.Exception) {
	return func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
		// Log the request
		data := frame.GetData()
		timestamp := time.Now().UTC().Format(time.RFC3339Nano)

		logEntry := map[string]interface{}{
			"timestamp":      timestamp,
			"protocol":       "modbus",
			"function_code":  funcCode,
			"function_name":  getFunctionName(funcCode),
			"data_length":    len(data),
			"raw_data":       fmt.Sprintf("%02x", data),
			"server_address": m.addr,
		}

		// Parse parameters based on function code
		switch funcCode {
		case 1, 2, 3, 4: // Read operations
			if len(data) >= 4 {
				startAddr := uint16(data[0])<<8 | uint16(data[1])
				quantity := uint16(data[2])<<8 | uint16(data[3])
				logEntry["start_address"] = startAddr
				logEntry["quantity"] = quantity
				logEntry["decoded_params"] = map[string]interface{}{
					"operation_type":   "read",
					"target_registers": fmt.Sprintf("%d to %d", startAddr, startAddr+quantity-1),
					"register_count":   quantity,
				}

				// For read operations, capture the register values after the operation
				defer func() {
					if funcCode == 3 || funcCode == 4 { // Holding or Input registers
						registerValues := make([]uint16, quantity)
						var sourceRegisters []uint16

						if funcCode == 3 { // Holding registers
							sourceRegisters = s.HoldingRegisters[:]
						} else { // Input registers
							sourceRegisters = s.InputRegisters[:]
						}

						// Extract the values that were read
						for i := uint16(0); i < quantity; i++ {
							if int(startAddr+i) < len(sourceRegisters) {
								registerValues[i] = sourceRegisters[startAddr+i]
							}
						}

						// Update register cache and check for complete dataset
						m.updateRegisterCache(startAddr, registerValues)

						// Log the actual values that were read
						log.Printf("Read Register Values - Start: %d, Count: %d, Values: %v",
							startAddr, quantity, registerValues)
					}
				}()
			}
		case 5, 6: // Write single operations
			if len(data) >= 4 {
				address := uint16(data[0])<<8 | uint16(data[1])
				value := uint16(data[2])<<8 | uint16(data[3])
				logEntry["address"] = address
				logEntry["value"] = value

				// Decode the value based on function code
				decodedValue := map[string]interface{}{
					"operation_type":  "write_single",
					"target_register": address,
					"raw_value":       value,
				}

				if funcCode == 5 { // Write single coil
					decodedValue["data_type"] = "coil"
					decodedValue["boolean_value"] = value != 0
					decodedValue["coil_state"] = map[string]interface{}{
						"on":  value != 0,
						"hex": fmt.Sprintf("0x%04X", value),
					}
				} else { // Write single register
					decodedValue["data_type"] = "register"
					decodedValue["decimal_value"] = value
					decodedValue["hex_value"] = fmt.Sprintf("0x%04X", value)
					decodedValue["binary_value"] = fmt.Sprintf("%016b", value)

					// Try to interpret as common industrial values
					if address <= 10 { // Common register range
						decodedValue["possible_interpretations"] = m.interpretRegisterValue(address, value)
					}
				}

				logEntry["decoded_params"] = decodedValue

				// Call OnWrite for write operations
				m.OnWrite(address, data[2:4])
			}
		case 15, 16: // Write multiple operations
			if len(data) >= 5 {
				startAddr := uint16(data[0])<<8 | uint16(data[1])
				quantity := uint16(data[2])<<8 | uint16(data[3])
				byteCount := data[4]
				logEntry["start_address"] = startAddr
				logEntry["quantity"] = quantity
				logEntry["byte_count"] = byteCount

				decodedValues := map[string]interface{}{
					"operation_type":   "write_multiple",
					"target_registers": fmt.Sprintf("%d to %d", startAddr, startAddr+quantity-1),
					"register_count":   quantity,
					"byte_count":       byteCount,
				}

				if len(data) >= int(5+byteCount) {
					values := data[5 : 5+byteCount]
					logEntry["values"] = fmt.Sprintf("%02x", values)

					// Decode multiple values based on function code
					if funcCode == 15 { // Write multiple coils
						decodedValues["data_type"] = "coils"
						coilStates := make([]map[string]interface{}, 0)

						for i := 0; i < int(quantity); i++ {
							byteIndex := i / 8
							bitIndex := i % 8
							if byteIndex < len(values) {
								bitValue := (values[byteIndex] >> bitIndex) & 1
								coilStates = append(coilStates, map[string]interface{}{
									"coil_address": startAddr + uint16(i),
									"state":        bitValue == 1,
									"bit_position": fmt.Sprintf("byte_%d_bit_%d", byteIndex, bitIndex),
								})
							}
						}
						decodedValues["coil_states"] = coilStates

					} else { // Write multiple registers
						decodedValues["data_type"] = "registers"
						registerValues := make([]map[string]interface{}, 0)

						for i := 0; i < len(values); i += 2 {
							if i+1 < len(values) {
								regValue := uint16(values[i])<<8 | uint16(values[i+1])
								regAddr := startAddr + uint16(i/2)

								regData := map[string]interface{}{
									"register_address": regAddr,
									"raw_value":        regValue,
									"decimal_value":    regValue,
									"hex_value":        fmt.Sprintf("0x%04X", regValue),
									"binary_value":     fmt.Sprintf("%016b", regValue),
								}

								// Add interpretations for common register addresses
								if regAddr <= 10 {
									regData["possible_interpretations"] = m.interpretRegisterValue(regAddr, regValue)
								}

								registerValues = append(registerValues, regData)
							}
						}
						decodedValues["register_values"] = registerValues
					}

					logEntry["decoded_params"] = decodedValues
					m.OnWrite(startAddr, data[5:5+byteCount])
				}
			}
		}

		logJSON, _ := json.Marshal(logEntry)
		log.Printf("Modbus Request: %s", string(logJSON))

		// Call the original handler
		return originalHandler(s, frame)
	}
}

func (m *ModbusServer) Stop() error {
	log.Println("Stopping Modbus server...")
	m.server.Close()
	return nil
}
