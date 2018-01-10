package main

import (
	"errors"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/hyperpilotio/node-agent/pkg/collector"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/processor"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"github.com/hyperpilotio/node-agent/pkg/snap"
)

type HyperpilotTask struct {
	Task      common.NodeTask
	Collector collector.Collector
	Processor processor.Processor
	Publisher map[string]*publisher.HyperpilotPublisher
}

func NewHyperpilotTask(
	task common.NodeTask,
	collector collector.Collector,
	processor processor.Processor,
	publishers map[string]*publisher.HyperpilotPublisher) (*HyperpilotTask, error) {

	return &HyperpilotTask{
		Task:      task,
		Collector: collector,
		Processor: processor,
		Publisher: publishers,
	}, nil
}

func (task *HyperpilotTask) Run(wg *sync.WaitGroup) {
	waitTime, _ := time.ParseDuration(task.Task.Schedule.Interval)
	tick := time.Tick(waitTime)
	go func() {
		for {
			select {
			case <-tick:
				metrics, _ := task.collect()
				if task.Processor != nil {
					metrics, _ = task.process(metrics)
				}
				for _, publish := range task.Publisher {
					publish.Put(metrics)
				}
			}
		}
		wg.Done()
	}()
}

func (task *HyperpilotTask) collect() ([]snap.Metric, error) {
	definition := task.Task
	pattern, err := glob.Compile(definition.Collect.Metrics)
	if err != nil {
		return nil, errors.New("Unable to compile collect namespace: " + err.Error())
	}

	metricTypes, err := task.Collector.GetMetricTypes(definition.Collect.Config)
	if err != nil {
		return nil, errors.New("Unable to get metric types: " + err.Error())
	}

	newMetricTypes := []snap.Metric{}
	for _, mts := range metricTypes {
		mts.Config = definition.Collect.Config
		if pattern.Match(mts.Namespace.String()) {
			newMetricTypes = append(newMetricTypes, mts)
		}
	}

	collectMetrics, err := task.Collector.CollectMetrics(newMetricTypes)
	if err != nil {
		return nil, errors.New("Unable to collect metric types: " + err.Error())
	}

	return collectMetrics, nil
}

func (task *HyperpilotTask) process(mts []snap.Metric) ([]snap.Metric, error) {
	return task.Processor.Process(mts, nil)
}
