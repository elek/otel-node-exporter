package nodeexporter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

var (
	typeStr = component.MustNewType("nodeexporter")
)

type Config struct {
	Interval string `mapstructure:"interval"`
}

func createDefaultConfig() component.Config {
	return &Config{}
}

// NewFactory creates a factory for tailtracer receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetrics, component.StabilityLevelAlpha))
}

func createMetrics(ctx context.Context, settings receiver.Settings, config component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return &NodeExporterReceiver{
		config:   config.(*Config),
		consumer: consumer,
		logger:   settings.Logger,
	}, nil
}
