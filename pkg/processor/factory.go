package processor

import (
	"errors"

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
		// TODO
		return nil, nil
	case "agent":
		// TODO
		return nil, nil
	default:
		return nil, errors.New("Unsupported processor type: " + name)
	}
}
