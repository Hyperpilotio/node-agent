package collector

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/collector/disk"
	"github.com/hyperpilotio/node-agent/pkg/collector/cpu"
	"github.com/hyperpilotio/node-agent/pkg/snap"
)

// Collector is a plugin which is the source of new data in the Snap pipeline.
type Collector interface {
	GetMetricTypes(snap.Config) ([]snap.Metric, error)
	CollectMetrics([]snap.Metric) ([]snap.Metric, error)
}

func NewCollector(name string) (Collector, error) {
	switch name {
	case "cpu":
		return cpu.New()
	case "disk":
		return disk.New()
		// case "docker":
		// case "prometheus":
		// case "psutil":
		// case "use":
	default:
		return nil, errors.New("Unsupported collector type: " + name)
	}
}
