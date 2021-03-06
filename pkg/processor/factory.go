package processor

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/processor/agent"
	"github.com/hyperpilotio/node-agent/pkg/processor/average"
	"github.com/hyperpilotio/node-agent/pkg/snap"
)

// Processor is a plugin which filters, aggregates, or decorates data in the
// Snap pipeline.
type Processor interface {
	Process([]snap.Metric, snap.Config) ([]snap.Metric, error)
}

func NewProcessor(name string) (Processor, error) {
	switch name {
	case "average":
		return avg.NewProcessor(), nil
	case "agent":
		return agent.NewProcessor(), nil
	default:
		return nil, errors.New("Unsupported processor type: " + name)
	}
}
