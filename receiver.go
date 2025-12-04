package nodeexporter

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/elek/otel-node-exporter/collector"
	"github.com/elek/otel-node-exporter/kingpin"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type NodeExporterReceiver struct {
	consumer consumer.Metrics
	cancel   context.CancelFunc
	config   *Config
	logger   *zap.Logger
	registry *prometheus.Registry
}

func (s *NodeExporterReceiver) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, s.cancel = context.WithCancel(ctx)

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	for key, value := range s.config.Flags {
		kingpin.Set(key, value)
	}

	collectors, err := collector.NewNodeCollector(log)
	if err != nil {
		s.logger.Error("failed to create uname collector", zap.Error(err))
	}
	s.registry = prometheus.NewRegistry()

	err = s.registry.Register(collectors)
	if err != nil {
		s.logger.Error("failed to register node collector", zap.Error(err))
	}

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
	dtos, err := s.registry.Gather()
	if err != nil {
		s.logger.Error("failed to gather metrics", zap.Error(err))
	}
	metrics := pmetric.NewMetrics()
	metric := metrics.ResourceMetrics().AppendEmpty()
	for _, dto := range dtos {
		switch *dto.Type {
		case io_prometheus_client.MetricType_GAUGE:
			sm := metric.ScopeMetrics().AppendEmpty()
			m := sm.Metrics().AppendEmpty()
			m.SetName(*dto.Name)
			if dto.Help != nil {
				m.SetDescription(*dto.Help)
			}

			for _, dtom := range dto.Metric {
				gauge := m.SetEmptyGauge()
				d := gauge.DataPoints().AppendEmpty()
				if dtom.Gauge.Value != nil {
					d.SetDoubleValue(*dtom.Gauge.Value)
				}
				for _, lp := range dtom.Label {
					d.Attributes().PutStr(lp.GetName(), lp.GetValue())
				}
				if dtom.TimestampMs == nil {
					d.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
				} else {
					d.SetTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(*dtom.TimestampMs)))
				}
			}
		case io_prometheus_client.MetricType_COUNTER:
			sm := metric.ScopeMetrics().AppendEmpty()
			m := sm.Metrics().AppendEmpty()
			m.SetName(*dto.Name)
			if dto.Help != nil {
				m.SetDescription(*dto.Help)
			}

			for _, dtom := range dto.Metric {
				counter := m.SetEmptySum()
				counter.SetIsMonotonic(true)
				counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				d := counter.DataPoints().AppendEmpty()
				if dtom.Counter.Value != nil {
					d.SetDoubleValue(*dtom.Counter.Value)
				}
				for _, lp := range dtom.Label {
					d.Attributes().PutStr(lp.GetName(), lp.GetValue())
				}
				if dtom.TimestampMs == nil {
					d.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
				} else {
					d.SetTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(*dtom.TimestampMs)))
				}
			}
		}

	}
	err = s.consumer.ConsumeMetrics(ctx, metrics)
	if err != nil {
		s.logger.Error("Failed to consume metrics", zap.Error(err))
	}
}
