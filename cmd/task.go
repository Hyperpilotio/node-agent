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
	"fmt"
)

type HyperpilotTask struct {
	Task      common.NodeTask
	Id        string
	Collector collector.Collector
	Processor processor.Processor
	Publisher map[string]*publisher.HyperpilotPublisher
}

func NewHyperpilotTask(
	task common.NodeTask,
	id string,
	collector collector.Collector,
	processor processor.Processor,
	publishers map[string]*publisher.HyperpilotPublisher) (*HyperpilotTask, error) {

	return &HyperpilotTask{
		Task:      task,
		Id:        id,
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

	var patterns []glob.Glob

	for name, _ := range definition.Collect.Metrics {
		pattern, err := glob.Compile(name)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to compile collect namespace {%s}: ", name, err.Error()))
		}

		patterns = append(patterns, pattern)

	}

	metricTypes, err := task.Collector.GetMetricTypes(definition.Collect.Config)
	if err != nil {
		return nil, errors.New("Unable to get metric types: " + err.Error())
	}

	newMetricTypes := []snap.Metric{}
	for _, mts := range metricTypes {
		mts.Config = definition.Collect.Config
		for _, pattern := range patterns {
			if pattern.Match(mts.Namespace.String()) {
				newMetricTypes = append(newMetricTypes, mts)
				break
			}
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
