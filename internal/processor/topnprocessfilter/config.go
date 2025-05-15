package topnprocessfilter

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	CPUThresholdPercent float64 `mapstructure:"cpu_threshold_percent"`

	MemoryThresholdPercent float64 `mapstructure:"memory_threshold_percent"`

	RetentionMinutes int64 `mapstructure:"retention_minutes"`

	StoragePath string `mapstructure:"storage_path"`
	
	EnableDynamicThresholds bool `mapstructure:"enable_dynamic_thresholds"`
}

func (cfg *Config) Validate() error {
	if cfg.CPUThresholdPercent < 0 || cfg.CPUThresholdPercent > 100 {
		return fmt.Errorf("cpu_threshold_percent must be between 0 and 100, got %v", cfg.CPUThresholdPercent)
	}
	if cfg.MemoryThresholdPercent < 0 || cfg.MemoryThresholdPercent > 100 {
		return fmt.Errorf("memory_threshold_percent must be between 0 and 100, got %v", cfg.MemoryThresholdPercent)
	}
	if cfg.RetentionMinutes <= 0 {
		return fmt.Errorf("retention_minutes must be greater than 0, got %v", cfg.RetentionMinutes)
	}
	return nil
}

var _ component.Config = (*Config)(nil)
