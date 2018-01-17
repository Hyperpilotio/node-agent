package publisher

import (
	"errors"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/hyperpilotio/node-agent/pkg/common"
	log "github.com/sirupsen/logrus"
)

type HyperpilotPublisher struct {
	MetricBuf     chan []snap.Metric
	Publisher     Publisher
	Config        snap.Config
	ErrReportChan chan common.TaskReport
}

func NewHyperpilotPublisher(pluginName string, config snap.Config, ch chan common.TaskReport) (*HyperpilotPublisher, error) {
	publisher, cfg, err := NewPublisher(pluginName, config)
	if err != nil {
		return nil, errors.New("Unable to create publisher: " + err.Error())
	}

	queue := make(chan []snap.Metric, 100)
	return &HyperpilotPublisher{
		MetricBuf:     queue,
		Publisher:     publisher,
		Config:        cfg,
		ErrReportChan: ch,
	}, nil
}

func (publisher *HyperpilotPublisher) Run() {
	go func() {
		for {
			select {
			case metrics := <-publisher.MetricBuf:
				if err := publisher.Publisher.Publish(metrics, publisher.Config); err != nil {
					publisher.reportError(common.TaskReport{
						LastErrorMsg:   err.Error(),
						LastErrorTime:  time.Now().UnixNano() / 1000000,
					})
					log.Warnf("Publiser push metric fail: %s", err.Error())
				}
			}
		}
	}()
}

func (publisher *HyperpilotPublisher) Put(metrics []snap.Metric) {
	publisher.MetricBuf <- metrics
}

func (publisher *HyperpilotPublisher) reportError(report common.TaskReport) {
	publisher.ErrReportChan <- report
}
