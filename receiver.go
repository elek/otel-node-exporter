package nodeexporter

import (
	"context"
	"log/slog"
	"time"

	"github.com/elek/otel-node-exporter/collector"
	"github.com/elek/otel-node-exporter/kingpin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type NodeExporterReceiver struct {
	consumer consumer.Metrics
	cancel   context.CancelFunc
	config   *Config
	logger   *zap.Logger
	producer *Producer
}

func (s *NodeExporterReceiver) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, s.cancel = context.WithCancel(ctx)

	sl := slog.New(zapslog.NewHandler(s.logger.Core()))

	for key, value := range s.config.Flags {
		kingpin.Set(key, value)
	}

	collectors, err := collector.NewNodeCollector(sl)
	if err != nil {
		s.logger.Error("failed to create uname collector", zap.Error(err))
	}
	registry := prometheus.NewRegistry()

	err = registry.Register(collectors)
	if err != nil {
		s.logger.Error("failed to register node collector", zap.Error(err))
	}

	s.producer = NewMetricProducer(WithGatherer(registry), WithScope("node_exporter"))
	interval, _ := time.ParseDuration(s.config.Interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.refresh(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (s *NodeExporterReceiver) Shutdown(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *NodeExporterReceiver) refresh(ctx context.Context) {
	metrics, err := s.producer.Produce(ctx)
	if err != nil {
		s.logger.Error("failed to gather metrics", zap.Error(err))
	}

	err = s.consumer.ConsumeMetrics(ctx, metrics)
	if err != nil {
		s.logger.Error("Failed to consume metrics", zap.Error(err))
	}
}
