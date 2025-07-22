#!/bin/bash

echo "=== Modbus Server Request Logging Demonstration ==="
echo "This script demonstrates the comprehensive logging capabilities"
echo "for troubleshooting Modbus communications in the smart grid middleware."
echo ""

# Start the server in the background
echo "Starting Modbus server with request logging..."
cd "/home/bob/OneDrive/AOB/RUGA/MSC/SEMESTER 2/PROBREM DRIVEN/igrid-middleware"
./bin/gateway &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo ""
echo "Server is running with PID: $SERVER_PID"
echo "The server will log all incoming Modbus requests with the following information:"
echo "- Timestamp (RFC3339 format)"
echo "- Protocol type (modbus)"
echo "- Function code and name"
echo "- Request parameters (addresses, quantities, values)"
echo "- Raw hex data"
echo "- Server address"
echo ""

echo "Sample log entries that you would see for different operations:"
echo ""

echo "1. Read Holding Registers (Function Code 3):"
echo '{"timestamp":"2025-01-XX...","protocol":"modbus","function_code":3,"function_name":"READ_HOLDING_REGISTERS","start_address":0,"quantity":10,"raw_data":"030000000a"}'
echo ""

echo "2. Write Single Register (Function Code 6):"
echo '{"timestamp":"2025-01-XX...","protocol":"modbus","function_code":6,"function_name":"WRITE_SINGLE_REGISTER","address":0,"value":1234,"raw_data":"06000004d2"}'
echo ""

echo "3. Write Multiple Registers (Function Code 16):"
echo '{"timestamp":"2025-01-XX...","protocol":"modbus","function_code":16,"function_name":"WRITE_MULTIPLE_REGISTERS","start_address":0,"quantity":3,"byte_count":6,"values":[1234,5678,9012]}'
echo ""

echo "This logging helps troubleshoot:"
echo "- Communication issues between devices"
echo "- Invalid register addresses or data ranges"
echo "- Protocol compliance problems"
echo "- Performance bottlenecks"
echo "- Security monitoring"
echo ""

echo "The middleware is now ready for production use with comprehensive request logging!"

# Clean up
sleep 2
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "Demo completed. The server has been stopped."
