package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/hyperpilotio/node-agent/pkg/analyzer"
	"github.com/hyperpilotio/node-agent/pkg/collector"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/processor"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

type PublishConfig struct {
	Publisher         []*HyperpilotPublisher
	AnalyzerPublisher []*HyperpilotPublisher
}

type HyperpilotTask struct {
	Task           *common.NodeTask
	Id             string
	Collector      collector.Collector
	Processor      processor.Processor
	Analyzer       analyzer.Analyzer
	PublishConfig  *PublishConfig
	CollectMetrics []snap.Metric
	FailureCount   int64
	Agent          *NodeAgent
}

func NewHyperpilotTask(
	task *common.NodeTask,
	id string,
	allMetricTypes []snap.Metric,
	collector collector.Collector,
	processor processor.Processor,
	analyzer analyzer.Analyzer,
	agent *NodeAgent) (*HyperpilotTask, error) {
	var pubs []*HyperpilotPublisher
	for _, pubId := range *task.Publish {
		p, ok := agent.Publishers[pubId]
		if ok {
			log.Infof("Publisher {%s} is loaded for Task {%s}", pubId, task.Id)
			pubs = append(pubs, p)
		} else {
			log.Warnf("Publisher {%s} is not loaded, skip", pubId)
		}
	}

	var analyzerPubs []*HyperpilotPublisher
	if task.Analyze != nil {
		for _, pubId := range *task.Analyze.Publish {
			p, ok := agent.Publishers[pubId]
			if ok {
				log.Infof("Publisher {%s} is loaded for Task {%s}", pubId, task.Id)
				analyzerPubs = append(analyzerPubs, p)
			} else {
				log.Warnf("Publisher {%s} is not loaded, skip", pubId)
			}
		}
	}

	metricPatterns := []glob.Glob{}
	for name := range task.Collect.Metrics {
		pattern, err := glob.Compile(name)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to compile collect namespace {%s}: ", name, err.Error()))
		}
		metricPatterns = append(metricPatterns, pattern)
	}

	cmts := getCollectMetricTypes(metricPatterns, allMetricTypes, task.Collect)
	if len(cmts) == 0 {
		errMsg := fmt.Sprintf("No metric match namespace for %s, no metrics are needed to collect", task.Id)
		log.Warnf(errMsg)
		return nil, errors.New(errMsg)
	}

	return &HyperpilotTask{
		Task:      task,
		Id:        id,
		Collector: collector,
		Processor: processor,
		Analyzer:  analyzer,
		PublishConfig: &PublishConfig{
			Publisher:         pubs,
			AnalyzerPublisher: analyzerPubs,
		},
		CollectMetrics: cmts,
		Agent:          agent,
	}, nil
}

func (task *HyperpilotTask) Run() {
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
				metrics, err := task.collect()
				if err != nil {
					task.FailureCount++
					log.Warnf("collect metric fail for %s, skip this time: %s", task.Task.Id, err.Error())
					task.reportError(err)
					continue
				}
				if task.Processor != nil {
					metrics, err = task.process(metrics, task.Task.Process.Config)
					if err != nil {
						task.FailureCount++
						task.reportError(err)
						log.Warnf("process metric fail for %s, skip this time: %s", task.Task.Id, err.Error())
						continue
					}
				}
				for _, publish := range task.PublishConfig.Publisher {
					publish.Put(metrics)
				}

				// Because analyze will be written to another database,
				// so the code as publish below, to avoid analyze error,
				// snap or snapaverage did not successfully write data
				if task.Analyzer != nil {
					derivedMetrics, err := task.analyze(metrics, task.Task.Analyze.Config)
					if err != nil {
						task.FailureCount++
						task.reportError(err)
						log.Warnf("analyze metric fail for %s, skip this time: %s", task.Task.Id, err.Error())
						continue
					}
					for _, publish := range task.PublishConfig.AnalyzerPublisher {
						if len(derivedMetrics) > 0 {
							publish.Put(derivedMetrics)
						}
					}
				}
			}
		}
	}()
}

func getCollectMetricTypes(
	metricPatterns []glob.Glob,
	allMetricTypes []snap.Metric,
	collect *common.Collect) []snap.Metric {
	collectMetricTypes := []snap.Metric{}
	for _, mt := range allMetricTypes {
		mt.Config = collect.Config
		namespace := mt.Namespace.String()
		matchNamespace := false
		for _, pattern := range metricPatterns {
			if pattern.Match(namespace) {
				collectMetricTypes = append(collectMetricTypes, mt)
				matchNamespace = true
				break
			}
		}

		if !matchNamespace {
			for name, _ := range collect.Metrics {
				if strings.HasPrefix(name, namespace) {
					collectMetricTypes = append(collectMetricTypes, mt)
					break
				}
			}
		}
	}

	return collectMetricTypes
}

func addTags(tags map[string]map[string]string, mts []snap.Metric) []snap.Metric {
	if len(tags) == 0 {
		return mts
	}

	newMts := []snap.Metric{}
	for _, mt := range mts {
		if mt.Tags == nil {
			mt.Tags = map[string]string{}
		}

		namespace := "/" + strings.Join(mt.Namespace.Strings(), "/")
		for prefix, entries := range tags {
			if strings.HasPrefix(namespace, prefix) {
				for k, v := range entries {
					mt.Tags[k] = v
				}
			}
		}
		newMts = append(newMts, mt)
	}

	return newMts
}

func (task *HyperpilotTask) collect() ([]snap.Metric, error) {
	collectMetrics, err := task.Collector.CollectMetrics(task.CollectMetrics)
	if err != nil {
		return nil, fmt.Errorf("Unable to collect metrics for %s: %s", task.Id, err.Error())
	}

	return addTags(task.Task.Collect.Tags, collectMetrics), nil
}

func (task *HyperpilotTask) process(mts []snap.Metric, cfg snap.Config) ([]snap.Metric, error) {
	return task.Processor.Process(mts, cfg)
}

func (task *HyperpilotTask) analyze(mts []snap.Metric, cfg snap.Config) ([]snap.Metric, error) {
	return task.Analyzer.Analyze(mts, cfg)
}

func (task *HyperpilotTask) reportError(err error) {
	report := common.TaskReport{
		Id:            task.Id,
		Plugin:        task.Task.Process.PluginName,
		LastErrorMsg:  err.Error(),
		LastErrorTime: time.Now().UnixNano() / 1000000,
		FailureCount:  task.FailureCount,
	}
	task.Agent.UpdateTaskReport(report)
}
