package modbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tbrandon/mbserver"
)

type ModbusServer struct {
	server *mbserver.Server
	addr   string
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

func NewServer() (*ModbusServer, error) {
	srv := mbserver.NewServer()

	modbusServer := &ModbusServer{
		server: srv,
		addr:   ":5020",
	}

	srv.HoldingRegisters[0] = 1234
	srv.HoldingRegisters[1] = 5678
	srv.HoldingRegisters[2] = 9012
	srv.InputRegisters[0] = 3456
	srv.InputRegisters[1] = 7890
	srv.Debug = true

	return modbusServer, nil
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
// based on common smart grid and industrial automation conventions
func (m *ModbusServer) interpretRegisterValue(address uint16, value uint16) map[string]interface{} {
	interpretations := make(map[string]interface{})

	// Common smart grid register interpretations
	switch address {
	case 0: // Often voltage
		interpretations["voltage_v"] = float64(value) / 10.0    // Common scaling: value/10 = volts
		interpretations["voltage_mv"] = float64(value)          // Alternative: millivolts
		interpretations["percentage"] = float64(value) / 655.35 // 0-65535 as 0-100%

	case 1: // Often current
		interpretations["current_a"] = float64(value) / 100.0     // Common scaling: value/100 = amperes
		interpretations["current_ma"] = float64(value)            // Alternative: milliamperes
		interpretations["power_factor"] = float64(value) / 1000.0 // 0-1000 as 0.0-1.0

	case 2: // Often power
		interpretations["power_w"] = float64(value)              // Watts
		interpretations["power_kw"] = float64(value) / 1000.0    // Kilowatts
		interpretations["frequency_hz"] = float64(value) / 100.0 // Frequency in Hz (scaled)

	case 3: // Often energy
		interpretations["energy_wh"] = float64(value)            // Watt-hours
		interpretations["energy_kwh"] = float64(value) / 1000.0  // Kilowatt-hours
		interpretations["temperature_c"] = float64(value) / 10.0 // Temperature in Celsius

	case 4: // Often frequency or temperature
		interpretations["frequency_hz"] = float64(value) / 100.0 // Frequency: 5000 = 50.00 Hz
		interpretations["temperature_c"] = float64(value) / 10.0 // Temperature: 250 = 25.0°C
		interpretations["rpm"] = float64(value)                  // Motor RPM

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
