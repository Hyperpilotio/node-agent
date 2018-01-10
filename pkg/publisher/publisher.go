package publisher

import (
	"log"

	"github.com/hyperpilotio/node-agent/pkg/snap"
)

type HyperpilotPublisher struct {
	MetricBuf chan []snap.Metric
	Publisher snap.Publisher
	Config    snap.Config
}

func NewHyperpilotPublisher(plugin string, config snap.Config) *HyperpilotPublisher {

	var publisher snap.Publisher
	switch plugin {
	case "snap-plugin-publisher-influxdb":
		// publisher = influxdb.NewInfluxPublisher()
	default:
		log.Printf("not support publisher plugin")

	}

	return &HyperpilotPublisher{
		MetricBuf: make(chan []snap.Metric),
		Publisher: publisher,
		Config:    config,
	}
}

func (publisher *HyperpilotPublisher) Run() {
	go func() {
		for {
			select {
			case metrics := <-publisher.MetricBuf:
				// todo
				log.Printf("%s", metrics)
				publisher.Publisher.Publish(metrics, publisher.Config)

			}
		}
	}()
}

func (publisher *HyperpilotPublisher) Put(metrics []snap.Metric) {
	publisher.MetricBuf <- metrics
}
