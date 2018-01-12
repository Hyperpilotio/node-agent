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
	Config        *common.TasksDefinition
	taskLock      sync.Mutex
	Tasks         map[string]*HyperpilotTask
	publisherLock sync.Mutex
	Publishers    map[string]*publisher.HyperpilotPublisher
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

	log.Infof("%d Tasks are configured to load: ", len(taskDef.Tasks))
	for _, task := range taskDef.Tasks {
		if task.Process != nil {
			log.Infof("Task {%s}: collect={%s}, process={%s}, publisher = %s", task.Id, task.Collect.PluginName, task.Process.PluginName, *task.Publish)
		} else {
			log.Infof("Task {%s}: collect={%s}, publisher = %s", task.Id, task.Collect.PluginName, *task.Publish)
		}
	}

	log.Infof("%d Puslisher are configured to load", len(taskDef.Publish))
	for _, p := range taskDef.Publish {
		log.Infof("Publisher id = {%s}, type = {%s}", p.Id, p.PluginName)
	}

	return &NodeAgent{
		Config:     taskDef,
		Tasks:      make(map[string]*HyperpilotTask),
		Publishers: make(map[string]*publisher.HyperpilotPublisher),
	}, nil
}

func (nodeAgent *NodeAgent) Init() error {
	// init publisher first
	for _, p := range nodeAgent.Config.Publish {
		if err := nodeAgent.CreatePublisher(p); err != nil {
			log.Errorf("unable to create publisher {%s}: %s", p.Id, err.Error())
			return err
		}
	}

	// init all tasks
	for _, task := range nodeAgent.Config.Tasks {
		if err := nodeAgent.CreateTask(task); err != nil {
			log.Errorf("unable to create task {%s}: %s", task.Id, err.Error())
			return err
		}
	}
	return nil
}

func (nodeAgent *NodeAgent) CreateTask(task *common.NodeTask) error {
	nodeAgent.taskLock.Lock()
	defer nodeAgent.taskLock.Unlock()

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

	newTask, err := NewHyperpilotTask(task, task.Id, taskCollector, taskProcessor, nodeAgent.Publishers)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to new agent task {%s}: %s", task.Id, err.Error()))
	}

	if _, ok := nodeAgent.Tasks[task.Id]; ok {
		log.Warnf("Task id {%s} is duplicated, skip this task", task.Id)
		return nil
	}
	nodeAgent.Tasks[task.Id] = newTask
	return nil
}

func (nodeAgent *NodeAgent) CreatePublisher(p *common.Publish) error {
	nodeAgent.publisherLock.Lock()
	defer nodeAgent.publisherLock.Unlock()

	hpPublisher, err := publisher.NewHyperpilotPublisher(p.PluginName, p.Config)
	if err != nil {
		return fmt.Errorf("unable to new publisher id={%s}, type={%s}: %s", p.Id, p.PluginName, err.Error())
	}
	hpPublisher.Run()

	if _, ok := nodeAgent.Publishers[p.Id]; ok {
		log.Warnf("Publisher id {%s}, (type {%s}) is duplicated, skip this publisher", p.Id, p.PluginName)
		return nil
	}
	nodeAgent.Publishers[p.Id] = hpPublisher
	return nil
}

func (nodeAgent *NodeAgent) Run(wg *sync.WaitGroup) {
	for _, task := range nodeAgent.Tasks {
		wg.Add(1)
		task.Run(wg)
	}
}
