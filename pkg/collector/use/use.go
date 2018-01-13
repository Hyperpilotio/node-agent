package use

import (
	"path/filepath"
	"regexp"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
)

const (
	waitTime = 10 * time.Millisecond
)

var (
	procPath     = "/proc"
	sysFsNetPath = "/sys/class/net"
	diskStatPath = filepath.Join(procPath, "diskstats")
	cpuStatPath  = filepath.Join(procPath, "stat")
	loadAvgPath  = filepath.Join(procPath, "loadavg")
	memInfoPath  = filepath.Join(procPath, "meminfo")
	vmStatPath   = filepath.Join(procPath, "vmstat")
	metricLabels = []string{
		"utilization",
		"saturation",
	}
)

// Use contains values of previous measurments
type Use struct {
	host string
}

// NewUseCollector returns Use struct
func New() (*Use, error) {
	return &Use{}, nil
}

// CollectMetrics returns Use metrics
func (u *Use) CollectMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	metrics := make([]snap.Metric, len(mts))
	cpure := regexp.MustCompile(`^/intel/use/compute/.*`)
	netre := regexp.MustCompile(`^/intel/use/network/.*`)
	storre := regexp.MustCompile(`^/intel/use/storage/.*`)
	memre := regexp.MustCompile(`^/intel/use/memory/.*`)

	for i, p := range mts {
		ns := p.Namespace.String()
		switch {
		case cpure.MatchString(ns):
			metric, err := u.computeStat(p.Namespace)
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric

		case netre.MatchString(ns):
			metric, err := u.networkStat(p.Namespace)
			if err != nil {
				return nil, err
			}

			metrics[i] = *metric
		case storre.MatchString(ns):
			metric, err := u.diskStat(p.Namespace)
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		case memre.MatchString(ns):
			metric, err := memStat(p.Namespace)
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		}
		tags, err := hostTags()

		if err == nil {
			metrics[i].Tags = tags
		}
		metrics[i].Timestamp = time.Now()

	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by use snap
func (u *Use) GetMetricTypes(_ snap.Config) ([]snap.Metric, error) {
	mts := []snap.Metric{}

	cpu, err := getCPUMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, cpu...)
	net, err := getNetIOCounterMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, net...)
	disk, err := getDiskMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, disk...)
	mem, err := getMemMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, mem...)

	return mts, nil
}
