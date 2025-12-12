package client

import (
	"fmt"

	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"
	"github.com/spf13/viper"
)

type Config struct {
	Temporal temporal.Config
}

// LoadConfig reads configuration from the specified file path using Viper
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
