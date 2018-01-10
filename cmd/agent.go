package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"github.com/hyperpilotio/node-agent/pkg/collector"
	"github.com/hyperpilotio/node-agent/pkg/processor"
)

type NodeAgent struct {
	Config *common.TasksDefinition
	Tasks  []*HyperpilotTask

	mutex sync.Mutex
}

func NewNodeAgent(taskFilePath string) (*NodeAgent, error) {
	b, err := ioutil.ReadFile(taskFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s dir: %s", taskFilePath, err.Error())
	}

	taskDef := &common.TasksDefinition{}
	if err := json.Unmarshal(b, taskDef); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal json to TasksDefinition: %s", err.Error())
	}

	return &NodeAgent{
		Config: taskDef,
	}, nil
}

func (nodeAgent *NodeAgent) Init() error {
	// init publisher first
	publishers := make(map[string]*publisher.HyperpilotPublisher)
	for _, p := range nodeAgent.Config.Publish {
		hpPublisher, err := publisher.NewHyperpilotPublisher(p.PluginName, p.Config)
		if err != nil {
			return errors.New("Unable to new hyperpilot publisher:" + err.Error())
		}
		publishers[p.Name] = hpPublisher
		hpPublisher.Run()
	}

	// init all tasks
	for _, task := range nodeAgent.Config.Tasks {
		collectName := task.Collect.PluginName
		taskCollector, err := collector.NewCollector(collectName)
		if err != nil {
			return fmt.Errorf("Unable to new collector for %s: %s", collectName, err.Error())
		}

		var taskProcessor processor.Processor
		if task.Process != nil {
			processName := task.Process.PluginName
			taskProcessor, err = processor.NewProcessor(processName)
			if err != nil {
				return fmt.Errorf("Unable to new processor for %s: %s", processName, err.Error())
			}
		}

		t, err := NewHyperpilotTask(task, taskCollector, taskProcessor, publishers)
		if err != nil {
			return errors.New("Unable to new hyperpilot snap task:" + err.Error())
		}
		nodeAgent.Tasks = append(nodeAgent.Tasks, t)
	}
	return nil
}

func (a *NodeAgent) Run(wg *sync.WaitGroup) {
	for _, task := range a.Tasks {
		wg.Add(1)
		task.Run(wg)
	}
}
