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

// HTTPConfig struct definition added to fix undefined error.
type HTTPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
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
