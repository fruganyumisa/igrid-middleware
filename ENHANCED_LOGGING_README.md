# Enhanced Modbus Server Logging - Human-Readable Data Decoding

## Overview
The Modbus server now provides comprehensive human-readable decoding of all received data values, making troubleshooting and monitoring much easier for smart grid applications.

## Enhanced Features

### 1. **Smart Value Interpretation**
The system automatically interprets register values based on common smart grid conventions:

- **Register 0 (Voltage)**: 2400 → 240.0V (scaled by /10), 2400mV, 3.66% of range
- **Register 1 (Current)**: 1500 → 15.00A (scaled by /100), 1500mA, power factor 1.5
- **Register 2 (Power)**: 5000 → 5000W, 5.0kW, 50.00Hz (frequency)
- **Register 3 (Energy)**: 10000 → 10000Wh, 10.0kWh, 1000°C (temperature)
- **Register 4+ (General)**: 1000 → 10.00Hz, 1000 RPM, 100°C

### 2. **Multiple Data Format Representations**
Each value is decoded into multiple formats:
```json
{
  "raw_value": 2400,
  "decimal_value": 2400,
  "hex_value": "0x0960",
  "binary_value": "0000100101100000"
}
```

### 3. **Status Bit Decoding**
Every register value is interpreted as potential status bits:
```json
{
  "status_bits": {
    "active_bits": ["maintenance_required", "overload", "door_open", "heater_on"],
    "active_count": 4,
    "bit_details": {
      "system_ready": false,
      "alarm_active": false,
      "maintenance_required": true,
      "overload": true,
      // ... all 16 bits decoded
    }
  }
}
```

### 4. **BCD (Binary Coded Decimal) Support**
Automatic detection and decoding of BCD-encoded values commonly used in industrial equipment.

### 5. **Operation-Specific Decoding**

#### Read Operations
```json
{
  "decoded_params": {
    "operation_type": "read",
    "target_registers": "0 to 2",
    "register_count": 3
  }
}
```

#### Write Single Operations
```json
{
  "decoded_params": {
    "operation_type": "write_single",
    "target_register": 0,
    "data_type": "register",
    "possible_interpretations": {
      "voltage_v": 240.0,
      "voltage_mv": 2400,
      "percentage": 3.66
    }
  }
}
```

#### Write Multiple Operations
```json
{
  "decoded_params": {
    "operation_type": "write_multiple",
    "register_values": [
      {
        "register_address": 2,
        "raw_value": 5000,
        "possible_interpretations": {
          "power_w": 5000,
          "power_kw": 5.0,
          "frequency_hz": 50.0
        }
      }
    ]
  }
}
```

## Sample Enhanced Log Entry

```json
{
  "timestamp": "2025-07-22T22:55:47.399875089Z",
  "protocol": "modbus",
  "function_code": 6,
  "function_name": "WRITE_SINGLE_REGISTER",
  "address": 0,
  "value": 2400,
  "data_length": 4,
  "raw_data": "00000960",
  "server_address": ":5020",
  "decoded_params": {
    "operation_type": "write_single",
    "target_register": 0,
    "data_type": "register",
    "raw_value": 2400,
    "decimal_value": 2400,
    "hex_value": "0x0960",
    "binary_value": "0000100101100000",
    "possible_interpretations": {
      "voltage_v": 240.0,
      "voltage_mv": 2400,
      "percentage": 3.662165255207141,
      "status_bits": {
        "active_bits": ["maintenance_required", "overload", "door_open", "heater_on"],
        "active_count": 4
      },
      "interpreted_at": "2025-07-22T22:55:47.3999448Z"
    }
  }
}
```

## Benefits for Troubleshooting

### 1. **Immediate Value Understanding**
- No need to manually convert hex values
- Automatic scaling for common industrial units
- Multiple interpretation possibilities

### 2. **Smart Grid Specific Interpretations**
- Voltage, current, power, energy readings
- Frequency monitoring
- Temperature and status monitoring

### 3. **Comprehensive Status Monitoring**
- Bit-level status decoding
- Equipment state tracking
- Alarm and fault condition identification

### 4. **Historical Analysis**
- JSON format perfect for log aggregation
- Easy parsing for analytics tools
- Timestamp precision for sequence analysis

## Usage Instructions

### Starting the Server
```bash
cd "/path/to/igrid-middleware"
go build -o bin/gateway cmd/gateway/main.go
./bin/gateway --modbus-server
```

### Example Client Test
```bash
go run test_enhanced_logging.go
```

## Production Recommendations

1. **Log Rotation**: Implement rotation due to detailed logging
2. **Filtering**: Consider filtering by register ranges for specific monitoring
3. **Alerting**: Set up alerts on specific decoded values (e.g., voltage out of range)
4. **Analytics**: Use the JSON structure for trend analysis and predictive maintenance

## Status
✅ **Fully Implemented and Tested**
- All Modbus function codes supported
- Human-readable value decoding working
- Smart grid specific interpretations
- Production-ready logging format
- Comprehensive test coverage

The enhanced logging system is now ready for production deployment in smart grid environments, providing unprecedented visibility into Modbus communications with automatic data interpretation.
