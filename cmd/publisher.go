package main

import (
	"errors"
	"time"
	"fmt"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"github.com/hyperpilotio/node-agent/pkg/common/queue"
	"github.com/cenkalti/backoff"
)

type HyperpilotPublisher struct {
	Queue        *queue.Queue
	Task         *common.Publish
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

	queueSize := agent.Config.GetInt("PublisherQueueSize")
	return &HyperpilotPublisher{
		Queue:     queue.NewCappedQueue(queueSize),
		Task:      p,
		Publisher: publisher,
		Config:    cfg,
		Id:        p.Id,
		Agent:     agent,
	}, nil
}

func (publisher *HyperpilotPublisher) Run() {
	retryTimeout, err := time.ParseDuration(publisher.Agent.Config.GetString("PublisherTimeOut"))
	if err != nil {
		log.Warnf("Parse PublisherTimeOut {%s} fail, use default interval 3 min in publisher {%s}",
			publisher.Agent.Config.GetString("PublisherTimeOut"), publisher.Id)
		retryTimeout = 3 * time.Minute
	}

	batchSize := publisher.Agent.Config.GetInt("PublisherBatchSize")
	if batchSize < 1 {
		log.Warnf("Batch Size {%d} is not feasible, use 1 instead", publisher.Agent.Config.GetInt("PublisherBatchSize"))
		batchSize = 1
	}

	go func() {
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 10 * time.Second
		b.MaxInterval = 1 * time.Minute
		b.MaxElapsedTime = retryTimeout

		for {
			if !publisher.Queue.Empty() {
				var batchMetrics []snap.Metric
				for i := 0; i < batchSize; i++ {
					metrics := publisher.Queue.Dequeue()
					if metrics == nil {
						log.Warnf("Publisher {%s} get nil metric, because element number inside of queue is less than batch size {%d}",
							publisher.Id, batchSize)
						continue
					}
					batchMetrics = append(batchMetrics, metrics.([]snap.Metric) ...)
				}

				retryPublish := func() error {
					return publisher.Publisher.Publish(batchMetrics, publisher.Config)
				}

				err := backoff.Retry(retryPublish, b)
				if err != nil {
					publisher.FailureCount++
					publisher.reportError(err)
					log.Warnf("Publisher {%s} push metric fail, %d metrics are dropped: %s", publisher.Id, len(batchMetrics), err.Error())
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
