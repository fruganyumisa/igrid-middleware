// Smart Grid Modbus Device Examples
// This file shows various smart grid devices and their typical Modbus register configurations

package examples

import (
	"github.com/fruganyumisa/igrid-middleware/internal/adapters/modbus"
)

// Power Quality Meter Configuration
func PowerQualityMeterConfig() modbus.Config {
	return modbus.Config{
		Address:      "192.168.1.100:502",
		PollInterval: 1000, // 1 second for power quality monitoring
		Registers: []modbus.Register{
			// Voltage measurements (3-phase)
			{Name: "voltage_l1_n", Address: 3000, Quantity: 2},  // Phase 1 to Neutral
			{Name: "voltage_l2_n", Address: 3002, Quantity: 2},  // Phase 2 to Neutral
			{Name: "voltage_l3_n", Address: 3004, Quantity: 2},  // Phase 3 to Neutral
			{Name: "voltage_l1_l2", Address: 3006, Quantity: 2}, // Line to Line voltages
			{Name: "voltage_l2_l3", Address: 3008, Quantity: 2},
			{Name: "voltage_l3_l1", Address: 3010, Quantity: 2},

			// Current measurements
			{Name: "current_l1", Address: 3012, Quantity: 2},
			{Name: "current_l2", Address: 3014, Quantity: 2},
			{Name: "current_l3", Address: 3016, Quantity: 2},
			{Name: "current_neutral", Address: 3018, Quantity: 2},

			// Power measurements
			{Name: "active_power_total", Address: 3020, Quantity: 2},
			{Name: "reactive_power_total", Address: 3022, Quantity: 2},
			{Name: "apparent_power_total", Address: 3024, Quantity: 2},
			{Name: "power_factor_total", Address: 3026, Quantity: 2},

			// Frequency and THD
			{Name: "frequency", Address: 3028, Quantity: 1},
			{Name: "thd_voltage_l1", Address: 3030, Quantity: 2}, // Total Harmonic Distortion
			{Name: "thd_current_l1", Address: 3032, Quantity: 2},
		},
	}
}

// Energy Meter Configuration
func EnergyMeterConfig() modbus.Config {
	return modbus.Config{
		Address:      "192.168.1.101:502",
		PollInterval: 30000, // 30 seconds for energy readings
		Registers: []modbus.Register{
			// Energy registers (typically 32-bit or 64-bit values)
			{Name: "energy_import_active", Address: 4000, Quantity: 2},   // kWh imported
			{Name: "energy_export_active", Address: 4002, Quantity: 2},   // kWh exported
			{Name: "energy_import_reactive", Address: 4004, Quantity: 2}, // kVArh imported
			{Name: "energy_export_reactive", Address: 4006, Quantity: 2}, // kVArh exported

			// Demand registers
			{Name: "demand_active_max", Address: 4008, Quantity: 2},     // Maximum demand
			{Name: "demand_active_current", Address: 4010, Quantity: 2}, // Current demand

			// Billing information
			{Name: "billing_period", Address: 4012, Quantity: 1},
			{Name: "billing_reset_count", Address: 4013, Quantity: 1},
		},
	}
}

// Protective Relay Configuration
func ProtectiveRelayConfig() modbus.Config {
	return modbus.Config{
		Address:      "192.168.1.102:502",
		PollInterval: 500, // 500ms for protection systems
		Registers: []modbus.Register{
			// Protection status
			{Name: "protection_status", Address: 5000, Quantity: 1},
			{Name: "fault_status", Address: 5001, Quantity: 1},
			{Name: "alarm_status", Address: 5002, Quantity: 1},

			// Trip information
			{Name: "last_trip_cause", Address: 5003, Quantity: 1},
			{Name: "trip_counter", Address: 5004, Quantity: 1},

			// Settings and thresholds
			{Name: "overcurrent_threshold", Address: 5010, Quantity: 2},
			{Name: "overvoltage_threshold", Address: 5012, Quantity: 2},
			{Name: "undervoltage_threshold", Address: 5014, Quantity: 2},
			{Name: "frequency_deviation_threshold", Address: 5016, Quantity: 2},

			// Real-time measurements for protection
			{Name: "fault_current_l1", Address: 5020, Quantity: 2},
			{Name: "fault_current_l2", Address: 5022, Quantity: 2},
			{Name: "fault_current_l3", Address: 5024, Quantity: 2},
		},
	}
}

// Weather Station Configuration
func WeatherStationConfig() modbus.Config {
	return modbus.Config{
		Address:      "192.168.1.103:502",
		PollInterval: 60000, // 1 minute for weather data
		Registers: []modbus.Register{
			// Environmental measurements
			{Name: "ambient_temperature", Address: 6000, Quantity: 1},
			{Name: "humidity", Address: 6001, Quantity: 1},
			{Name: "wind_speed", Address: 6002, Quantity: 2},
			{Name: "wind_direction", Address: 6004, Quantity: 1},
			{Name: "solar_irradiance", Address: 6005, Quantity: 2},
			{Name: "atmospheric_pressure", Address: 6007, Quantity: 2},
			{Name: "rainfall", Address: 6009, Quantity: 2},

			// Equipment temperature monitoring
			{Name: "transformer_oil_temp", Address: 6020, Quantity: 1},
			{Name: "switchgear_temp", Address: 6021, Quantity: 1},
			{Name: "cable_temp", Address: 6022, Quantity: 1},
		},
	}
}

// Battery Energy Storage System (BESS) Configuration
func BatteryStorageConfig() modbus.Config {
	return modbus.Config{
		Address:      "192.168.1.104:502",
		PollInterval: 2000, // 2 seconds for battery monitoring
		Registers: []modbus.Register{
			// Battery status
			{Name: "battery_state_of_charge", Address: 7000, Quantity: 1}, // SOC %
			{Name: "battery_state_of_health", Address: 7001, Quantity: 1}, // SOH %
			{Name: "battery_voltage", Address: 7002, Quantity: 2},         // DC voltage
			{Name: "battery_current", Address: 7004, Quantity: 2},         // DC current (+ charging, - discharging)
			{Name: "battery_temperature", Address: 7006, Quantity: 1},

			// Power conversion
			{Name: "inverter_ac_power", Address: 7010, Quantity: 2}, // AC power output/input
			{Name: "inverter_efficiency", Address: 7012, Quantity: 1},
			{Name: "inverter_status", Address: 7013, Quantity: 1},

			// Control and limits
			{Name: "charge_rate_limit", Address: 7020, Quantity: 2},
			{Name: "discharge_rate_limit", Address: 7022, Quantity: 2},
			{Name: "operational_mode", Address: 7024, Quantity: 1}, // 0=Off, 1=Charge, 2=Discharge, 3=Standby
		},
	}
}

// Example payload structures for different devices
const (
	PowerQualityPayloadExample = `{
		"voltage_l1_n": 230.5,
		"voltage_l2_n": 229.8,
		"voltage_l3_n": 231.2,
		"current_l1": 15.75,
		"current_l2": 16.20,
		"current_l3": 15.95,
		"active_power_total": 11250.5,
		"reactive_power_total": 2150.3,
		"power_factor_total": 0.982,
		"frequency": 50.02,
		"thd_voltage_l1": 2.1,
		"thd_current_l1": 4.5
	}`

	EnergyMeterPayloadExample = `{
		"energy_import_active": 123456.789,
		"energy_export_active": 5432.123,
		"energy_import_reactive": 45678.901,
		"demand_active_max": 125.5,
		"demand_active_current": 87.3,
		"billing_period": 202507
	}`

	ProtectiveRelayPayloadExample = `{
		"protection_status": 1,
		"fault_status": 0,
		"alarm_status": 0,
		"last_trip_cause": 0,
		"trip_counter": 15,
		"overcurrent_threshold": 200.0,
		"overvoltage_threshold": 250.0,
		"undervoltage_threshold": 200.0
	}`

	WeatherStationPayloadExample = `{
		"ambient_temperature": 25.5,
		"humidity": 65.2,
		"wind_speed": 12.5,
		"wind_direction": 225,
		"solar_irradiance": 850.3,
		"atmospheric_pressure": 1013.25,
		"transformer_oil_temp": 45.8
	}`

	BatteryStoragePayloadExample = `{
		"battery_state_of_charge": 85,
		"battery_state_of_health": 95,
		"battery_voltage": 650.5,
		"battery_current": -25.3,
		"battery_temperature": 32.1,
		"inverter_ac_power": 15750.0,
		"inverter_efficiency": 96.5,
		"operational_mode": 2
	}`
)
