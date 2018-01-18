package main

import (
	"errors"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"github.com/hyperpilotio/node-agent/pkg/common/queue"
	"github.com/cenkalti/backoff"
)

type HyperpilotPublisher struct {
	Queue        *queue.Queue
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

	return &HyperpilotPublisher{
		Queue:     queue.NewCappedQueue(100),
		Publisher: publisher,
		Config:    cfg,
		Id:        p.Id,
		Agent:     agent,
	}, nil
}

func (publisher *HyperpilotPublisher) Run() {
	go func() {
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 10 * time.Second
		b.MaxInterval = 1 * time.Minute
		b.MaxElapsedTime = 3 * time.Minute

		for {
			if !publisher.Queue.Empty() {
				//todo: get batch
				metrics := publisher.Queue.Dequeue()
				if metrics == nil {
					log.Warnf("Publisher {%s} push nil metric, should not happen", publisher.Id)
					continue
				}

				retryPublish := func() error {
					return publisher.Publisher.Publish(metrics.([]snap.Metric), publisher.Config)
				}

				err := backoff.Retry(retryPublish, b)
				if err != nil {
					publisher.FailureCount++
					publisher.reportError(common.PublisherReport{
						Id:            publisher.Id,
						LastErrorMsg:  err.Error(),
						LastErrorTime: time.Now().UnixNano() / 1000000,
						FailureCount:  publisher.FailureCount,
					})
					log.Warnf("Publisher {%s} push metric fail, %d metrics are dropped: %s", publisher.Id, len(metrics.([]snap.Metric)), err.Error())
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (publisher *HyperpilotPublisher) Put(metrics []snap.Metric) {
	for !publisher.Queue.Enqueue(metrics) {
		log.Warnf("Enqueue fail due to full queue, remove oldest metric")
		publisher.Queue.Dequeue()
	}
}

func (publisher *HyperpilotPublisher) reportError(report common.PublisherReport) {
	publisher.Agent.UpdatePublishReport(report)
}
