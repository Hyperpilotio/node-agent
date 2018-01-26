package analyzer

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/analyzer/nodeanalyzer"
	"github.com/hyperpilotio/node-agent/pkg/snap"
)

// Analyzer is a plugin which filters, aggregates, or decorates data in the
// Snap pipeline.
type Analyzer interface {
	Analyze([]snap.Metric, snap.Config) ([]snap.Metric, error)
}

func NewAnalyzer(name string) (Analyzer, error) {
	switch name {
	case "nodeanalyzer":
		return nodeanalyzer.NewAnalyzer(), nil
	default:
		return nil, errors.New("Unsupported analyzer type: " + name)
	}
}
