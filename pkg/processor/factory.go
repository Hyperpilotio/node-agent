package processor

import (
	"errors"
	"os"

	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/processor/average"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

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
