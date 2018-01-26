package common

import "github.com/hyperpilotio/node-agent/pkg/snap"

type Schedule struct {
	Interval string `json:"interval"`
}

type metricInfo struct {
	Version_ int `json:"version"yaml:"version"`
}

type Collect struct {
	PluginName string                       `json:"plugin"`
	Metrics    map[string]metricInfo        `json:"metrics"`
	Config     snap.Config                  `json:"config"`
	Tags       map[string]map[string]string `json:"tags,omitempty"`
}

type Process struct {
	PluginName string      `json:"plugin"`
	Config     snap.Config `json:"config"`
}

type Analyze struct {
	PluginName string      `json:"plugin"`
	Config     snap.Config `json:"config"`
	Publish    *[]string   `json:"publish"`
}

type Publish struct {
	PluginName string      `json:"plugin"`
	Id         string      `json:"id"`
	Config     snap.Config `json:"config"`
}

type NodeTask struct {
	Id       string    `json:"id"`
	Schedule Schedule  `json:"schedule"`
	Collect  *Collect  `json:"collect"`
	Process  *Process  `json:"process"`
	Analyze  *Analyze  `json:"analyze"`
	Publish  *[]string `json:"publish"`
}

type TasksDefinition struct {
	Tasks   []*NodeTask `json:"tasks"`
	Publish []*Publish  `json:"publish"`
}
