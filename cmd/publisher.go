package main

import (
	"errors"
	"time"
	"fmt"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
)

type HyperpilotPublisher struct {
	Task         *common.Publish
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
		return nil, errors.New(fmt.Sprintf("Unable to create publisher {%s}: %s", p.PluginName, err.Error()))
	}

	queue := make(chan []snap.Metric, 100)
	return &HyperpilotPublisher{
		Task:      p,
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
					publisher.reportError(err)
					log.Warnf("Publiser push metric fail: %s", err.Error())
				}
			}
		}
	}()
}

func (publisher *HyperpilotPublisher) Put(metrics []snap.Metric) {
	publisher.MetricBuf <- metrics
}

func (publisher *HyperpilotPublisher) reportError(err error) {
	report := common.PublisherReport{
		Id:            publisher.Id,
		Plugin:        publisher.Task.PluginName,
		LastErrorMsg:  err.Error(),
		LastErrorTime: time.Now().UnixNano() / 1000000,
		FailureCount:  publisher.FailureCount,
	}
	publisher.Agent.UpdatePublishReport(report)
}
