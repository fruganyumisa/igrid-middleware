package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging       LoggingConfig            `yaml:"logging"`
	Modbus        ModbusConfig             `yaml:"modbus"`
	DNP3          DNP3Config               `yaml:"dnp3"`
	MQTT          MQTTConfig               `yaml:"mqtt"`
	HTTP          HTTPConfig               `yaml:"http"`
	Validation    ValidationConfig         `yaml:"validation"`
	Normalization map[string]MappingConfig `yaml:"normalization"`
	SchemaPath    string                   `yaml:"schema_path"`
}

type MappingConfig struct {
	JSONPath    string  `yaml:"json_path"`
	Scaling     float64 `yaml:"scaling"`
	DataType    string  `yaml:"data_type"`
	Unit        string  `yaml:"unit"`
	SourceField string  `yaml:"source_field"`
	TargetField string  `yaml:"target_field"`
	Transform   string  `yaml:"transform,omitempty"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type ModbusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	UnitID  byte   `yaml:"unitId"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

type MQTTConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BrokerHost   string `yaml:"brokerHost"`
	BrokerPort   int    `yaml:"brokerPort"`
	ClientID     string `yaml:"clientId"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PublishTopic string `yaml:"publishTopic"`
	QoS          byte   `yaml:"qos"`
}

// DNP3Config struct definition added to fix undefined error.
type DNP3Config struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout int    `yaml:"timeout"`
}

// HTTPConfig struct definition for HTTP client and DMS endpoint configuration
type HTTPConfig struct {
	Enabled     bool              `yaml:"enabled"`
	Host        string            `yaml:"host"`
	Port        int               `yaml:"port"`
	DMS         DMSConfig         `yaml:"dms"`         // DMS endpoint configuration
	Timeout     int               `yaml:"timeout"`     // HTTP timeout in seconds
	Retries     int               `yaml:"retries"`     // Number of retry attempts
	Headers     map[string]string `yaml:"headers"`     // Custom HTTP headers
}

// DMSConfig contains DMS (Distribution Management System) endpoint configuration
type DMSConfig struct {
	Enabled     bool              `yaml:"enabled"`     // Enable DMS integration
	BaseURL     string            `yaml:"base_url"`    // DMS base URL (e.g., "https://dms.example.com")
	Endpoint    string            `yaml:"endpoint"`    // Data endpoint (e.g., "/api/v1/devices/data")
	Timeout     int               `yaml:"timeout"`     // Request timeout in seconds
	AuthType    string            `yaml:"auth_type"`   // Authentication type: "bearer", "basic", "apikey", "none"
	AuthToken   string            `yaml:"auth_token"`  // Bearer token or API key
	Username    string            `yaml:"username"`    // For basic auth
	Password    string            `yaml:"password"`    // For basic auth
	Headers     map[string]string `yaml:"headers"`     // Additional headers
	RetryConfig RetryConfig       `yaml:"retry"`       // Retry configuration
}

// RetryConfig defines retry behavior for HTTP requests
type RetryConfig struct {
	Enabled     bool `yaml:"enabled"`      // Enable retry mechanism
	MaxAttempts int  `yaml:"max_attempts"` // Maximum retry attempts
	DelayMs     int  `yaml:"delay_ms"`     // Delay between retries in milliseconds
}

// ValidationConfig struct definition added to fix undefined error.
type ValidationConfig struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
}

// Other config types omitted for brevity...

type NormalizationConfig struct {
	Enabled bool   `yaml:"enabled"`
	Method  string `yaml:"method"`
}

func Load(filename string) (*Config, error) {
	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
