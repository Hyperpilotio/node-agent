package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hyperpilotio/node-agent/pkg/analyzer"
	"github.com/hyperpilotio/node-agent/pkg/collector"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/processor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type NodeAgent struct {
	Config              *viper.Viper
	TasksDef            *common.TasksDefinition
	taskLock            sync.Mutex
	Tasks               map[string]*HyperpilotTask
	publisherLock       sync.Mutex
	Publishers          map[string]*HyperpilotPublisher
	taskReportLock      sync.RWMutex
	TasksReport         map[string]common.TaskReport
	publisherReportLock sync.RWMutex
	PublishersReport    map[string]common.PublisherReport
}

func NewNodeAgent(config *viper.Viper) (*NodeAgent, error) {
	taskFilePath := config.GetString("TaskConfiguration")
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
		taskInfos := []string{}
		taskInfos = append(taskInfos, fmt.Sprintf("Task {%s}: collect={%s}", task.Id, task.Collect.PluginName))
		if task.Process != nil {
			taskInfos = append(taskInfos, fmt.Sprintf("process={%s}", task.Process.PluginName))
		}

		if task.Analyze != nil {
			taskInfos = append(taskInfos, fmt.Sprintf("analyze={%s}", task.Analyze.PluginName))
			if task.Analyze.Publish != nil {
				taskInfos = append(taskInfos, fmt.Sprintf("analyzePublisher={%s}", *task.Analyze.Publish))
			}
		}

		taskInfos = append(taskInfos, fmt.Sprintf("publisher = %s", *task.Publish))
		log.Infof(strings.Join(taskInfos, ", "))
	}

	log.Infof("%d Puslisher are configured to load", len(taskDef.Publish))
	for _, p := range taskDef.Publish {
		log.Infof("Publisher id = {%s}, type = {%s}", p.Id, p.PluginName)
	}

	return &NodeAgent{
		Config:           config,
		TasksDef:         taskDef,
		Tasks:            make(map[string]*HyperpilotTask),
		Publishers:       make(map[string]*HyperpilotPublisher),
		TasksReport:      make(map[string]common.TaskReport),
		PublishersReport: make(map[string]common.PublisherReport),
	}, nil
}

func (nodeAgent *NodeAgent) Init() error {
	// init publisher first
	for _, p := range nodeAgent.TasksDef.Publish {
		if err := nodeAgent.CreatePublisher(p); err != nil {
			log.Errorf("unable to create publisher {%s}: %s", p.Id, err.Error())
			return err
		}
	}

	// init all tasks
	for _, task := range nodeAgent.TasksDef.Tasks {
		if err := nodeAgent.CreateTask(task); err != nil {
			log.Errorf("Unable to create task {%s}: %s", task.Id, err.Error())
			return err
		}
	}

	// start node agent api server
	go nodeAgent.startAPIServer(nodeAgent.Config.GetString("APIServerPort"))

	return nil
}

func (nodeAgent *NodeAgent) CreateTask(task *common.NodeTask) error {
	nodeAgent.taskLock.Lock()
	defer nodeAgent.taskLock.Unlock()

	collectName := task.Collect.PluginName
	taskCollector, err := collector.NewCollector(collectName)
	if err != nil {
		return fmt.Errorf("Unable to new %s collector for task %s: %s", collectName, task.Id, err.Error())
	}

	metricTypes, err := taskCollector.GetMetricTypes(task.Collect.Config)
	if err != nil {
		return fmt.Errorf("Unable to get %s metric types: %s", collectName, err.Error())
	}

	var taskProcessor processor.Processor
	if task.Process != nil {
		processName := task.Process.PluginName
		taskProcessor, err = processor.NewProcessor(processName)
		if err != nil {
			return fmt.Errorf("unable to new %s processor for task %s: %s", processName, task.Id, err.Error())
		}
	}

	var taskAnalyzer analyzer.Analyzer
	if task.Analyze != nil {
		analyzeName := task.Analyze.PluginName
		taskAnalyzer, err = analyzer.NewAnalyzer(analyzeName)
		if err != nil {
			return fmt.Errorf("unable to new %s analyzer for task %s: %s", analyzeName, task.Id, err.Error())
		}
	}

	newTask, err := NewHyperpilotTask(task, task.Id, metricTypes,
		taskCollector, taskProcessor, taskAnalyzer, nodeAgent)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to new agent task {%s}: %s", task.Id, err.Error()))
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

	hpPublisher, err := NewHyperpilotPublisher(nodeAgent, p)
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

func (nodeAgent *NodeAgent) Run() {
	for _, task := range nodeAgent.Tasks {
		task.Run()
	}
}

func (nodeAgent *NodeAgent) startAPIServer(port string) {
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	clusterGroup := router.Group("/")
	{
		clusterGroup.GET("/report", nodeAgent.Report)
	}

	log.Infof("API Server starts")
	err := router.Run(":" + port)
	if err != nil {
		log.Errorf(" api server cannot start :%s", err.Error())
	}
}

func (nodeAgent *NodeAgent) UpdateTaskReport(report common.TaskReport) {
	nodeAgent.taskReportLock.Lock()
	defer nodeAgent.taskReportLock.Unlock()

	id := report.Id
	report.Id = ""
	nodeAgent.TasksReport[id] = report
}

func (nodeAgent *NodeAgent) UpdatePublishReport(report common.PublisherReport) {
	nodeAgent.publisherReportLock.Lock()
	defer nodeAgent.publisherReportLock.Unlock()

	id := report.Id
	report.Id = ""
	nodeAgent.PublishersReport[id] = report
}

func (nodeAgent *NodeAgent) Report(c *gin.Context) {
	report := common.Report{}
	fillTaskReport(nodeAgent, &report)
	fillPublisherReport(nodeAgent, &report)

	c.JSON(http.StatusOK, report)
}

func fillTaskReport(agent *NodeAgent, report *common.Report) {
	agent.taskReportLock.RLock()
	defer agent.taskReportLock.RUnlock()

	for key, value := range agent.TasksReport {
		report.Tasks[key] = value
	}
}

func fillPublisherReport(agent *NodeAgent, report *common.Report) {
	agent.publisherReportLock.RLock()
	defer agent.publisherReportLock.RUnlock()

	for key, value := range agent.PublishersReport {
		report.Publisher[key] = value
	}
}
