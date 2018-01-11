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
	log "github.com/sirupsen/logrus"
)

type NodeAgent struct {
	Config *common.TasksDefinition
	Tasks  map[string]*HyperpilotTask
	mutex  sync.Mutex
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

	log.Printf("%d Tasks are configured to load: ", len(taskDef.Tasks))
	for _, task := range taskDef.Tasks {
		if task.Process != nil {
			log.Printf("Task {%s}: collect={%s}, process={%s}, publisher = %s", task.Id, task.Collect.PluginName, task.Process.PluginName, *task.Publish)
		} else {
			log.Printf("Task {%s}: collect={%s}, publisher = %s", task.Id, task.Collect.PluginName, *task.Publish)
		}
	}

	log.Printf("%d Puslisher are configured to load", len(taskDef.Publish))
	for _, p := range taskDef.Publish {
		log.Printf("name = {%s}, type = {%s}", p.Id, p.PluginName)
	}

	return &NodeAgent{
		Config: taskDef,
		Tasks:  make(map[string]*HyperpilotTask),
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
		publishers[p.Id] = hpPublisher
		hpPublisher.Run()
	}

	// init all tasks
	for _, task := range nodeAgent.Config.Tasks {
		collectName := task.Collect.PluginName
		taskCollector, err := collector.NewCollector(collectName)
		if err != nil {
			return fmt.Errorf("unable to new %s collector for task %s: %s", collectName, task.Id, err.Error())
		}

		var taskProcessor processor.Processor
		if task.Process != nil {
			processName := task.Process.PluginName
			taskProcessor, err = processor.NewProcessor(processName)
			if err != nil {
				return fmt.Errorf("unable to new %s processor for task %s: %s", processName, task.Id, err.Error())
			}
		}

		newTask, err := NewHyperpilotTask(task, task.Id, taskCollector, taskProcessor, publishers)
		if err != nil {
			return errors.New(fmt.Sprintf("unable to new agent task {%s}: %s", task.Id, err.Error()))
		}
		nodeAgent.mutex.Lock()
		if _, ok := nodeAgent.Tasks[task.Id]; ok {
			log.Warnf("Task id {%s} is duplicated, skip this task", task.Id)
			nodeAgent.mutex.Unlock()
			continue
		}
		nodeAgent.Tasks[task.Id] = newTask
		nodeAgent.mutex.Unlock()

	}
	return nil
}

func (a *NodeAgent) Run(wg *sync.WaitGroup) {
	for _, task := range a.Tasks {
		wg.Add(1)
		task.Run(wg)
	}
}
