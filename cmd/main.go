package main

import (
	"sync"
	log "github.com/sirupsen/logrus"
)

func main() {
	path := "/etc/node_agent/disk-task-test.json"
	//path := "./conf/disk-task-test.json"

	nodeAgent, err := NewNodeAgent(path)
	if err != nil {
		log.Printf("create agent fail: %s", err.Error())
		return
	}

	if err := nodeAgent.Init(); err != nil {
		log.Printf("Agent Init() fail: %s", err.Error())
	}

	// use wg to wait
	taskWg := &sync.WaitGroup{}
	nodeAgent.Run(taskWg)
	taskWg.Wait()
}
