package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

const (
	Name    = "file"
	Version = 3
)

type filePublisher struct {
}

type MetricToPublish struct {
	// The timestamp from when the metric was created.
	Timestamp time.Time         `json:"timestamp"`
	Namespace string            `json:"namespace"`
	Data      interface{}       `json:"data"`
	Unit      string            `json:"unit"`
	Tags      map[string]string `json:"tags"`
	Version   int64             `json:"version"`
}

//New returns an instance of filePublisher
func New() *filePublisher {
	return &filePublisher{}
}

func (f *filePublisher) Publish(mts []snap.Metric, cfg snap.Config) error {
	logger := log.New()
	logger.Debug("Publishing started")

	destination, err := cfg.GetString("file")
	if err != nil {
		return fmt.Errorf("%s: %s", err, "file")
	}

	logger.Debugf("Publishing %v metrics to %s", len(mts), destination)
	file, err := os.OpenFile(destination, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Error opening file: %v", err)
	}
	w := bufio.NewWriter(file)

	// format metrics types to metrics to be published
	metrics := formatMetricTypes(mts)
	jsonOut, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("Error while marshalling metrics to JSON: %v", err)
	}

	w.Write(jsonOut)
	w.WriteString("\n")
	w.Flush()

	return nil
}

// formatMetricTypes returns metrics in format to be publish as a JSON based on incoming metrics types;
// i.a. namespace is formatted as a single string
func formatMetricTypes(mts []snap.Metric) []MetricToPublish {
	var metrics []MetricToPublish
	for _, mt := range mts {
		metrics = append(metrics, MetricToPublish{
			Timestamp: mt.Timestamp,
			Namespace: mt.Namespace.String(),
			Data:      mt.Data,
			Unit:      mt.Unit,
			Tags:      mt.Tags,
			Version:   mt.Version,
		})
	}
	return metrics
}
