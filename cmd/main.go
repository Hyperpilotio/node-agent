package main

import (
	"github.com/hyperpilotio/node-agent/pkg/collector"
	"sync"
)

func main() {
	path := "./conf/tasks.json"

	agent, _ := collector.NewAgent(path)
	agent.Init()


	// use wg to wait

	taskWg := &sync.WaitGroup{}
	agent.Run(taskWg)
	taskWg.Wait()

}
