package nodeexporter

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/node_exporter/collector"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type NodeExporterReceiver struct {
	consumer   consumer.Metrics
	cancel     context.CancelFunc
	config     *Config
	logger     *zap.Logger
	collectors []collector.Collector
}

func (s *NodeExporterReceiver) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, s.cancel = context.WithCancel(ctx)

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	unameCollector, err := collector.NewUnameCollector(log)
	if err != nil {
		s.logger.Error("failed to create uname collector", zap.Error(err))
	}
	s.collectors = append(s.collectors, unameCollector)

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
	for _, component := range s.collectors {
		cname := fmt.Sprintf("%T", component)
		s.logger.Info("Scraping metrics from component", zap.String("component", cname))
		ch := make(chan prometheus.Metric)
		go func() {
			err := component.Update(ch)
			if err != nil {
				s.logger.Error("failed to scrape metrics", zap.Error(err))
			}
			close(ch)
		}()

		metrics := pmetric.NewMetrics()
		metric := metrics.ResourceMetrics().AppendEmpty()

	processingloop:
		for {

			var dt dto.Metric
			select {
			case mtr, ok := <-ch:
				if !ok {
					break processingloop
				}

				err := mtr.Write(&dt)
				if err != nil {
					s.logger.Error("failed to write metrics", zap.Error(err))
					continue
				}

				sm := metric.ScopeMetrics().AppendEmpty()
				sm.Scope().SetName(cname)
				m := sm.Metrics().AppendEmpty()

				if dt.Gauge != nil {
					gauge := m.SetEmptyGauge()
					d := gauge.DataPoints().AppendEmpty()
					if dt.Gauge.Value != nil {
						d.SetIntValue(int64(*dt.Gauge.Value))
					}
					for _, lp := range dt.Label {
						d.Attributes().PutStr(lp.GetName(), lp.GetValue())
					}
					if dt.TimestampMs != nil {
						d.SetTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(*dt.TimestampMs)))
					}
				}

			case <-ctx.Done():
				break processingloop
			}
		}
		err := s.consumer.ConsumeMetrics(ctx, metrics)
		if err != nil {
			s.logger.Error("Failed to consume metrics", zap.Error(err))
		}
	}
}
