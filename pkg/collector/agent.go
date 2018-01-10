package collector

import (
	"github.com/hyperpilotio/node-agent/pkg/common"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/hyperpilotio/node-agent/pkg/publisher"
	"sync"
)

type Agent struct {
	Config *common.TasksDefinition
	Tasks  []*HyperpilotTask
}

func NewAgent(taskFilePath string) (*Agent, error) {

	var agent Agent

	b, err := ioutil.ReadFile(taskFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s dir: %s", taskFilePath, err.Error())
	}

	taskDef := common.TasksDefinition{}
	if err := json.Unmarshal(b, &taskDef); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal json to TasksDefinition: %s", err.Error())
	}

	agent.Config = &taskDef

	return &agent, nil
}

func (a *Agent) Init() error {

	// init publisher first
	publishers := make(map[string]*publisher.HyperpilotPublisher)
	for _, p := range a.Config.Publish {
		pub := publisher.NewHyperpilotPublisher(p.PluginName, p.Config)
		publishers[p.Name] = pub
	}

	// init all tasks
	for _, task := range a.Config.Tasks {
		t := NewHyperpilotTask(task, publishers)
		a.Tasks = append(a.Tasks, t)
	}
	return nil
}

func (a *Agent) Run(wg *sync.WaitGroup) {
	for _, task := range a.Tasks {
		wg.Add(1)
		task.Run(wg)
	}
}
