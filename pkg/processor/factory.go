package processor

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/hyperpilotio/node-agent/pkg/processor/average"
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
	default:
		return nil, errors.New("Unsupported processor type: " + name)
	}
}
