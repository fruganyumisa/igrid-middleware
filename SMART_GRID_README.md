# Smart Grid Data Processing System

## Overview

This enhanced iGrid Middleware now supports comprehensive smart grid data processing with automatic normalization and routing to DMS (Distribution Management System) applications. The system receives Modbus data, interprets it as smart grid measurements, and automatically routes normalized data to external systems.

## Enhanced Features

### 🔌 Smart Grid Data Support

The system now supports the following smart grid parameters:

| Register | Parameter | Scaling | Example | Unit |
|----------|-----------|---------|---------|------|
| 0 | Voltage | /100 | 26247 → 262.47V | V |
| 1 | Current | /1000 | 14560 → 14.56A | A |
| 2 | Active Power | /10 | 15210 → 1521.0W | W |
| 3 | Reactive Power | /100 | 12280 → 122.80VAR | VAR |
| 4 | Frequency | /100 | 5034 → 50.34Hz | Hz |
| 5 | Breaker Status | Raw | 0 = Open, 1 = Closed | - |
| 6 | Tap Position | Raw | 22 | - |
| 7-8 | Energy | 32-bit | Combined registers | Wh |
| 9 | Temperature | /100 | 6370 → 63.70°C | °C |
| 10 | Humidity | /100 | 6048 → 60.48% | % |
| 11 | Alarm | Raw | 0 = No alarm | - |
| 12-13 | Timestamp | 32-bit | Unix timestamp | - |
| 14 | Device ID | Raw | Device identifier | - |

### 🚀 Automatic Data Routing

When a complete dataset (registers 0-14) is read via Modbus, the system:

1. **Interprets** raw register values using smart grid scaling
2. **Normalizes** data into structured JSON format
3. **Routes** to DMS application with filtered payload:

```json
{
    "deviceId": "SMART_GRID_DEVICE_001",
    "voltage": 262.47,
    "current": 14.56,
    "temperature": 63.70
}
```

### 📊 Enhanced Logging

All Modbus operations are logged with:
- Human-readable data interpretations
- Multiple data format representations (hex, binary, decimal)
- Status bit decoding
- Automatic scaling and unit conversions

## Usage Examples

### Running the Smart Grid Test Server

```bash
# Build and run the test server
go build ./cmd/smart-grid-test
./smart-grid-test
```

The server will start on port 5020 with pre-loaded smart grid data.

### Testing with Modbus Client

```bash
# Build and run the test client (server must be running)
go build ./cmd/modbus-client-test
./modbus-client-test
```

### Expected Output

**Server Log (when client reads data):**
```
Modbus Request: {"timestamp":"2025-07-23T03:17:25Z","protocol":"modbus","function_code":3,"function_name":"READ_HOLDING_REGISTERS","data_length":4,"raw_data":"000000f","start_address":0,"quantity":15}

Complete smart grid dataset detected, processing...
Extracted Smart Grid Data: {"deviceId":"SMART_GRID_METER_001","voltage":262.47,"current":14.56,"active_power":1521,"temperature":63.7,"timestamp":1753227100,"received_at":"2025-07-23T03:17:25Z"}

🚀 DMS Payload: {"deviceId":"SMART_GRID_METER_001","voltage":262.47,"current":14.56,"temperature":63.7}
```

**Client Output:**
```
📈 Raw Register Values:
   Register  0: 26247 (0x6687) - Voltage: 262.47V
   Register  1: 14560 (0x38E0) - Current: 14.560A
   Register  2: 15210 (0x3B6A) - Active Power: 1521.0W
   Register  9:  6370 (0x18E2) - Temperature: 63.70°C
```

## Integration Architecture

```
[Modbus Device] → [iGrid Middleware] → [DMS Application]
                       ↓
                 [Data Processing]
                       ↓
               [Human-readable Logs]
```

### Data Flow

1. **Input**: Raw Modbus register data from smart grid devices
2. **Processing**: Automatic scaling, interpretation, and normalization
3. **Output**: Filtered JSON payload sent to DMS via HTTP POST
4. **Logging**: Comprehensive troubleshooting logs with interpretations

## Configuration

### Modbus Server Configuration

The enhanced server can be configured with:
- Device ID for identification
- Custom data routing implementation
- HTTP endpoint configuration for DMS communication

### Example Integration

```go
// Create HTTP client for DMS
httpClient := http.NewHTTPClient(http.HTTPClientConfig{
    BaseURL: "https://dms.example.com",
    Timeout: 30 * time.Second,
}, logger)

// Create router
router := http.NewHTTPRouter(httpClient)

// Create Modbus server with routing
server, err := modbus.NewServer(router, "DEVICE_001")
```

## Benefits

✅ **Automatic Data Interpretation**: No manual scaling or conversion needed
✅ **Production-Ready Logging**: Immediate understanding of data values
✅ **Flexible Routing**: Easy integration with multiple DMS systems
✅ **Real-time Processing**: Automatic detection and routing of complete datasets
✅ **Troubleshooting Friendly**: Human-readable logs for debugging

## Troubleshooting

### Common Issues

1. **No data routing**: Ensure all registers 0-14 are read in sequence
2. **Incorrect scaling**: Verify register interpretations match your device
3. **DMS connection**: Check HTTP client configuration and endpoint availability

### Debug Logging

The system provides detailed logs for:
- Raw Modbus data (hex format)
- Interpreted values with units
- DMS payload generation
- HTTP request/response details

This enhanced system makes smart grid data integration significantly easier by providing automatic interpretation, normalization, and routing capabilities.
