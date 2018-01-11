package publisher

import (
	"errors"

	"github.com/hyperpilotio/node-agent/pkg/snap"
)

type HyperpilotPublisher struct {
	MetricBuf chan []snap.Metric
	Publisher Publisher
	Config    snap.Config
}

func NewHyperpilotPublisher(pluginName string, config snap.Config) (*HyperpilotPublisher, error) {
	publisher, cfg, err := NewPublisher(pluginName, config)
	if err != nil {
		return nil, errors.New("Unable to create publisher: " + err.Error())
	}

	queue := make(chan []snap.Metric, 100)
	return &HyperpilotPublisher{
		MetricBuf: queue,
		Publisher: publisher,
		Config:    cfg,
	}, nil
}

func (publisher *HyperpilotPublisher) Run() {
	go func() {
		for {
			select {
			case metrics := <-publisher.MetricBuf:
				// TODO
				publisher.Publisher.Publish(metrics, publisher.Config)
			}
		}
	}()
}

func (publisher *HyperpilotPublisher) Put(metrics []snap.Metric) {
	publisher.MetricBuf <- metrics
}
