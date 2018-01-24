package ddagent

import (
	"errors"
	"encoding/json"
	"strings"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/hyperpilotio/node-agent/pkg/collector/ddagent/protocol"
	log "github.com/sirupsen/logrus"
)

type DdAgent struct {
	tcp        protocol.Receiver
	inMetric   chan *snap.Metric
	outMetrics chan []snap.Metric
	done       chan struct{}
	isStarted  bool
}

func New() (*DdAgent, error) {
	return &DdAgent{
		tcp:        protocol.NewTCPListener(8888),
		inMetric:   make(chan *snap.Metric, 1000),
		outMetrics: make(chan []snap.Metric, 1000),
		done:       make(chan struct{}),
		isStarted:  false,
	}, nil
}

func (d *DdAgent) GetMetricTypes(cfg snap.Config) ([]snap.Metric, error) {
	mts := []snap.Metric{}
	vals := []string{"ddagent"}
	for _, val := range vals {
		metric := snap.Metric{
			Namespace: snap.NewNamespace(val),
		}
		mts = append(mts, metric)
	}

	return mts, nil
}

func (d *DdAgent) StreamMetrics(mts []snap.Metric) error {
	if d.isStarted {
		return errors.New("server already started")
	}
	log.Info("Starting tcp to receive datadog metric")
	if err := d.tcp.Start(); err != nil {
		return err
	}
	d.isStarted = true

	go func() {
		log.Infof("start goroutine to parse datadog metric")
		for {
			select {
			case data := <-d.tcp.Data():
				metric, err := parseData(data, mts)
				if err != nil {
					log.Warnf(err.Error())
					continue
				}
				d.inMetric <- metric
			case <-d.done:
				break
			}
		}
	}()

	// routine that dispatches statsd metrics to all available streams
	go func() {
		log.Infof("start goroutine to transform snap metric")
		for {
			select {
			case m := <-d.inMetric:
				d.outMetrics <- []snap.Metric{*m}
			case <-d.done:
				return
			}
		}
	}()
	return nil
}

func (d *DdAgent) Metrics() chan []snap.Metric {
	return d.outMetrics
}

func parseData(data []byte, mts []snap.Metric) (*snap.Metric, error) {
	var ddMetric Metric
	var snapMetric snap.Metric

	if err := json.Unmarshal(data, &ddMetric); err != nil {
		log.Errorf("Unmarshal receiving datadog metric {%s} failed: %s", string(data), err.Error())
		return nil, err
	}

	ns := snap.NewNamespace("ddagent")
	ns = ns.AddStaticElements(strings.Split(ddMetric.MetricName, ".")...)

	//todo: compare mts
	//return nil, errors.New("")

	//todo: fill in tag from datadog

	snapMetric = snap.Metric{
		Namespace: ns,
		Timestamp: time.Unix(ddMetric.Timestamp, 0),
		Data:      ddMetric.Value,
	}
	return &snapMetric, nil
}
