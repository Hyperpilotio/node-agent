package publisher

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/publisher/file"
	"github.com/hyperpilotio/node-agent/pkg/publisher/influxdb"
	"github.com/hyperpilotio/node-agent/pkg/snap"
)

// Publisher is a sink in the Snap pipeline.  It publishes data into another
// System, completing a Workflow path.
type Publisher interface {
	Publish([]snap.Metric, snap.Config) error
}

func NewPublisher(name string, cfg snap.Config) (Publisher, snap.Config, error) {
	switch name {
	case "file":
		return file.New(), cfg, nil
	case "influxdb":
		newCfg := cfg
		newCfg["port"] = int64(cfg["port"].(float64))
		return influxdb.NewInfluxPublisher(), newCfg, nil
	default:
		return nil, nil, errors.New("Unsupported publisher type: " + name)
	}
}
