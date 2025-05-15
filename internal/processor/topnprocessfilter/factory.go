package topnprocessfilter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr = "topnprocessfilter"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, component.StabilityLevelBeta),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		CPUThresholdPercent:    defaultCPUThresholdPercent,
		MemoryThresholdPercent: defaultMemoryThresholdPercent,
		RetentionMinutes:       defaultRetentionMinutes,
		StoragePath:            defaultStoragePath,
	}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	pCfg := cfg.(*Config)
	proc, err := newProcessor(set.Logger, pCfg, nextConsumer)
	if err != nil {
		return nil, err
	}

	return proc, nil
}
