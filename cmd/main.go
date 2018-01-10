package main

import (
	"sync"
)

func main() {
	path := "./conf/tasks.json"

	nodeAgent, _ := NewNodeAgent(path)
	nodeAgent.Init()

	// use wg to wait
	taskWg := &sync.WaitGroup{}
	nodeAgent.Run(taskWg)
	taskWg.Wait()
}
