package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ModbusConfig struct {
	Address      string `yaml:"address"`
	PollInterval int    `yaml:"poll_interval"`
	// ... other fields ...
}

type MQTTConfig struct {
	BrokerURL   string `yaml:"broker_url"`
	ClientID    string `yaml:"client_id"`
	TLSEnabled  bool   `yaml:"tls_enabled"`
	CertPath    string `yaml:"cert_path"`
	KeyPath     string `yaml:"key_path"`
	TopicPrefix string `yaml:"topic_prefix"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	// ... other fields ...
	// Add other fields as needed
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type Config struct {
	Logging LoggingConfig `yaml:"logging"`
	Modbus  ModbusConfig  `yaml:"modbus"`
	MQTT    MQTTConfig    `yaml:"mqtt"`
	// ... other protocol configs ...
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
