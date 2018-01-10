package collector

import (
	snap "github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"time"
	"errors"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/gobwas/glob"
	"github.com/hyperpilotio/snap-plugin-collector-prometheus/prometheus"
	"github.com/hyperpilotio/snap-average-counter-processor/agent"

	"github.com/intelsdi-x/snap-plugin-collector-docker/collector"


	"sync"
	"log"
)

type HyperpilotTask struct {
	Task      common.NodeTask
	Collector snap.Collector
	Processor snap.Processor
	Publisher map[string]*publisher.HyperpilotPublisher
}

func NewHyperpilotTask(task common.NodeTask, publishers map[string]*publisher.HyperpilotPublisher) *HyperpilotTask {
	collector := newCollector(task.Collect.PluginName)
	processor := newProcessor(task.Process.PluginName)

	return &HyperpilotTask{
		Publisher: publishers,
		Task:      task,
		Processor: processor,
		Collector: collector,
	}
}

func newCollector(plugin string) snap.Collector {

	var collector snap.Collector

	switch plugin {
	case "snap-plugin-collector-prometheus":
		collector = prometheus.New()
	//case "snap-plugin-collector-docker":

	default:
		return nil
	}
	return collector
}

func newProcessor(plugin string) snap.Processor {
	switch plugin {
	case "snap-average-counter-processor":
		return agent.NewProcessor()
	default:
		return nil
	}
	return nil
}

func (task *HyperpilotTask) Run(wg *sync.WaitGroup) {

	tick := time.Tick(5 * time.Second)

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

	pattern, _ := glob.Compile(definition.Collect.Metrics)

	log.Print(task.Collector)
	log.Print(definition.Collect.Config)

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
