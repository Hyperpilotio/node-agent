package common

import "github.com/hyperpilotio/node-agent/pkg/snap"

type Schedule struct {
	Interval string `json:"interval"`
}

type Collect struct {
	PluginName string      `json:"plugin"`
	Metrics    string      `json:"metrics"`
	Config     snap.Config `json:"config"`
	Tags       snap.Config `json:"tags"`
}

type Process struct {
	PluginName string      `json:"plugin"`
	Config     snap.Config `json:"config"`
}

type Publish struct {
	PluginName string      `json:"plugin"`
	Name       string      `json:"name"`
	Config     snap.Config `json:"config"`
}

type NodeTask struct {
	Schedule Schedule  `json:"schedule"`
	Collect  *Collect  `json:"collect"`
	Process  *Process  `json:"process"`
	Publish  *[]string `json:"publish"`
}

type TasksDefinition struct {
	Tasks   []NodeTask `json:"tasks"`
	Publish []*Publish `json:"publish"`
}
