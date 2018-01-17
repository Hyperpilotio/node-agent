package main

import (
	"errors"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
)

type HyperpilotPublisher struct {
	MetricBuf    chan []snap.Metric
	Publisher    publisher.Publisher
	Config       snap.Config
	Agent        *NodeAgent
	Id           string
	FailureCount int64
}

func NewHyperpilotPublisher(agent *NodeAgent, p *common.Publish) (*HyperpilotPublisher, error) {
	publisher, cfg, err := publisher.NewPublisher(p.PluginName, p.Config)
	if err != nil {
		return nil, errors.New("Unable to create publisher: " + err.Error())
	}

	queue := make(chan []snap.Metric, 100)
	return &HyperpilotPublisher{
		MetricBuf: queue,
		Publisher: publisher,
		Config:    cfg,
		Id:        p.Id,
		Agent:     agent,
	}, nil
}

func (publisher *HyperpilotPublisher) Run() {
	go func() {
		for {
			select {
			case metrics := <-publisher.MetricBuf:
				if err := publisher.Publisher.Publish(metrics, publisher.Config); err != nil {
					publisher.FailureCount++
					publisher.reportError(common.PublisherReport{
						Id:            publisher.Id,
						LastErrorMsg:  err.Error(),
						LastErrorTime: time.Now().UnixNano() / 1000000,
						FailureCount:  publisher.FailureCount,
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

func (publisher *HyperpilotPublisher) reportError(report common.PublisherReport) {
	publisher.Agent.UpdatePublishReport(report)
}
