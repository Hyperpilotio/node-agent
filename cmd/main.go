package main

import (
	"sync"
	"flag"

	log "github.com/sirupsen/logrus"
)

func main() {

	configPath := flag.String("config", "", "The file path to a config file")
	flag.Parse()

	if *configPath == "" {
		*configPath = "/etc/node_agent/tasks.json"
	}

	nodeAgent, err := NewNodeAgent(*configPath)
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
	log.Printf("Node Agent start running...")
	taskWg.Wait()
}
