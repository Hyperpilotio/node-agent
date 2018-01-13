package psutil

import (
	"fmt"
	"strings"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

func New() (*Psutil, error) {
	return &Psutil{}, nil
}

type Psutil struct {
}

// CollectMetrics returns metrics from gopsutil
func (p *Psutil) CollectMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	loadReqs := []snap.Namespace{}
	cpuReqs := []snap.Namespace{}
	memReqs := []snap.Namespace{}
	netReqs := []snap.Namespace{}
	diskReqs := []snap.Namespace{}

	for _, m := range mts {
		ns := m.Namespace
		switch ns[2].Value {
		case "load":
			loadReqs = append(loadReqs, ns)
		case "cpu":
			cpuReqs = append(cpuReqs, ns)
		case "vm":
			memReqs = append(memReqs, ns)
		case "net":
			netReqs = append(netReqs, ns)
		case "disk":
			diskReqs = append(diskReqs, ns)
		default:
			return nil, fmt.Errorf("Requested metric %s does not match any known psutil metric", m.Namespace.String())
		}
	}

	metrics := []snap.Metric{}

	loadMts, err := loadAvg(loadReqs)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, loadMts...)

	cpuMts, err := cpuTimes(cpuReqs)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, cpuMts...)

	memMts, err := virtualMemory(memReqs)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, memMts...)

	netMts, err := netIOCounters(netReqs)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, netMts...)
	mounts := getMountpoints(mts[0].Config)
	diskMts, err := getDiskUsageMetrics(diskReqs, mounts)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, diskMts...)

	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by gopsutil
func (p *Psutil) GetMetricTypes(_ snap.Config) ([]snap.Metric, error) {
	mts := []snap.Metric{}

	mts = append(mts, getLoadAvgMetricTypes()...)
	mts_, err := getCPUTimesMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, mts_...)
	mts = append(mts, getVirtualMemoryMetricTypes()...)

	mts_, err = getNetIOCounterMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, mts_...)
	mts = append(mts, getDiskUsageMetricTypes()...)

	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicy
// func (p *Psutil) GetConfigPolicy() (snap.ConfigPolicy, error) {
// 	c := snap.NewConfigPolicy()
// 	c.AddNewStringRule([]string{"intel", "psutil", "disk"},
// 		"mount_points", false)
// 	return *c, nil
// }

func getMountpoints(cfg snap.Config) []string {
	if mp, err := cfg.GetString("mount_points"); err == nil {
		if mp == "*" {
			return []string{"all"}
		}
		mountPoints := strings.Split(mp, "|")
		return mountPoints
	}
	return []string{"physical"}
}

type label struct {
	description string
	unit        string
}

func timeSpent(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debugf("%s took %s", name, elapsed)
}
