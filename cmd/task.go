package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/hyperpilotio/node-agent/pkg/collector"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/processor"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

type HyperpilotTask struct {
	Task           *common.NodeTask
	Id             string
	Collector      collector.Collector
	Processor      processor.Processor
	Publisher      []*publisher.HyperpilotPublisher
	MetricPatterns []glob.Glob
	CusTags       map[string]string
}

func NewHyperpilotTask(
	task *common.NodeTask,
	id string,
	collector collector.Collector,
	processor processor.Processor,
	publishers map[string]*publisher.HyperpilotPublisher) (*HyperpilotTask, error) {

	var pubs []*publisher.HyperpilotPublisher

	for _, pubId := range *task.Publish {
		p, ok := publishers[pubId]
		if ok {
			log.Infof("Publisher {%s} is loaded for Task {%s}", pubId, task.Id)
			pubs = append(pubs, p)
		} else {
			log.Warnf("Publisher {%s} is not loaded, skip", pubId)
		}
	}

	userCustTags := make(map[string]string)
	for _, entries := range task.Collect.Tags {
		for k, v := range entries {
			userCustTags[k] = v
		}
	}

	hypterpilotTask := HyperpilotTask{
		Task:      task,
		Id:        id,
		Collector: collector,
		Processor: processor,
		Publisher: pubs,
		CusTags:  userCustTags,
	}

	for name := range task.Collect.Metrics {
		pattern, err := glob.Compile(name)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to compile collect namespace {%s}: ", name, err.Error()))
		}
		hypterpilotTask.MetricPatterns = append(hypterpilotTask.MetricPatterns, pattern)
	}

	return &hypterpilotTask, nil
}

func (task *HyperpilotTask) Run(wg *sync.WaitGroup) {
	waitTime, err := time.ParseDuration(task.Task.Schedule.Interval)
	if err != nil {
		log.Warnf("Parse schedule interval {%s} fail, use default interval 5 seconds",
			task.Task.Schedule.Interval, err.Error())
		waitTime = 5 * time.Second
	}
	tick := time.Tick(waitTime)
	go func() {
		for {
			select {
			case <-tick:
				metrics, _ := task.collect()
				if task.Processor != nil {
					metrics, _ = task.process(metrics, task.Task.Process.Config)
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

	metricTypes, err := task.Collector.GetMetricTypes(definition.Collect.Config)
	if err != nil {
		return nil, errors.New("Unable to get metric types: " + err.Error())
	}

	newMetricTypes := []snap.Metric{}
	for _, mt := range metricTypes {
		mt.Config = definition.Collect.Config
		namespace := mt.Namespace.String()
		if mt.Tags == nil {
			mt.Tags = make(map[string]string)
		}

		for k, v := range task.CusTags {
			mt.Tags[k] = v
		}

		for _, pattern := range task.MetricPatterns {
			if pattern.Match(namespace) {
				newMetricTypes = append(newMetricTypes, mt)
				break
			}
		}
		for name, _ := range definition.Collect.Metrics {
			if strings.HasPrefix(name, namespace) {
				newMetricTypes = append(newMetricTypes, mt)
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

func (task *HyperpilotTask) process(mts []snap.Metric, cfg snap.Config) ([]snap.Metric, error) {
	return task.Processor.Process(mts, cfg)
}
