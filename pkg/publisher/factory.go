package publisher

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/publisher/file"
	"github.com/hyperpilotio/node-agent/pkg/snap"
)

// Publisher is a sink in the Snap pipeline.  It publishes data into another
// System, completing a Workflow path.
type Publisher interface {
	Publish([]snap.Metric, snap.Config) error
}

func NewPublisher(name string) (Publisher, error) {
	switch name {
	// case "cpu":
	case "file":
		return file.New(), nil
	// case "docker":
	// case "prometheus":
	// case "psutil":
	// case "use":
	default:
		return nil, errors.New("Unsupported publisher type: " + name)
	}
}