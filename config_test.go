package nodeexporter

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/elek/otel-node-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	nodeCollector, err := collector.NewNodeCollector(log)
	require.NoError(t, err)

	registry := prometheus.NewRegistry()
	err = registry.Register(nodeCollector)
	require.NoError(t, err)
	dtos, err := registry.Gather()
	for _, dto := range dtos {
		fmt.Println(*dto.Name, dto.Type)
		for _, m := range dto.Metric {
			if m.Gauge != nil {
				fmt.Println(" ", m.Label, *m.Gauge.Value)
			}
		}
	}

}
