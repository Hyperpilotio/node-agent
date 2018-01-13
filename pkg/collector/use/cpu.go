package use

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"strconv"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/shirou/gopsutil/cpu"
)

// CPUStat contains values of CPU previous measurments
type CPUStat struct {
	last    map[string]int64
	current map[string]int64
}

// LoadAvg struct with Host Load Statistics
type LoadAvg struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

// Utilization returns utilization of CPU
func (c *CPUStat) Utilization() (float64, error) {
	var err error

	c.last, err = readCPUStat()
	if err != nil {
		return 0.0, err
	}
	time.Sleep(waitTime)
	c.current, err = readCPUStat()
	if err != nil {
		return 0.0, err
	}
	deltaIdle := c.Idle(true) - c.Idle(false)
	deltaNonIdle := c.NonIdle(true) - c.NonIdle(false)
	if deltaIdle == 0.0 || deltaNonIdle == 0.0 {
		return 0.0, nil
	}

	return 100.00 * (deltaNonIdle / (deltaIdle + deltaNonIdle)), nil

}

// Idle returns current or last Idle time
func (c *CPUStat) Idle(actual bool) float64 {
	if actual {
		return float64(c.current["idle"])
	}
	return float64(c.last["idle"])
}

// NonIdle returns current or last NonIdle time
func (c *CPUStat) NonIdle(actual bool) float64 {
	if actual {
		return float64(c.current["user"] + c.current["nice"] + c.current["system"])
	}
	return float64(c.last["user"] + c.last["nice"] + c.last["system"])
}

func (p *Use) computeStat(ns snap.Namespace) (*snap.Metric, error) {

	switch {
	case regexp.MustCompile(`^/intel/use/compute/utilization`).MatchString(ns.String()):
		cpuStat := CPUStat{}
		metric, err := cpuStat.Utilization()
		if err != nil {
			return nil, err
		}
		return &snap.Metric{
			Namespace: ns,
			Data:      metric,
		}, nil
	case regexp.MustCompile(`^/intel/use/compute/saturation`).MatchString(ns.String()):
		metric, err := getSaturation()
		if err != nil {
			return nil, err
		}
		return &snap.Metric{
			Namespace: ns,
			Data:      metric,
		}, nil
	}

	return nil, fmt.Errorf("Unknown error processing %v", ns)
}

func getCPUMetricTypes() ([]snap.Metric, error) {
	var mts []snap.Metric
	for _, name := range metricLabels {
		mts = append(mts, snap.Metric{Namespace: snap.NewNamespace("intel", "use", "compute", name)})
	}
	return mts, nil
}

func getSaturation() (float64, error) {
	cpus, err := cpu.Times(true)
	if err != nil {
		return 0, err
	}
	cpuCount := len(cpus)
	load, err := readLoad()
	if err != nil {
		return 0, err
	}
	return load.Load1 / float64(cpuCount), nil
}

func readLoad() (*LoadAvg, error) {
	filename := loadAvgPath
	lines, err := readLines(filename)
	load := &LoadAvg{}
	if err != nil {
		return load, err
	}
	fields := strings.Fields(lines[0])
	load.Load1, err = strconv.ParseFloat((fields[0]), 64)
	load.Load5, err = strconv.ParseFloat((fields[1]), 64)
	load.Load15, err = strconv.ParseFloat((fields[2]), 64)
	if err != nil {
		return nil, err
	}

	return load, nil
}

func readCPUStat() (map[string]int64, error) {
	content, err := readLines(cpuStatPath)
	if err != nil {
		return nil, err
	}

	CPUStat := strings.Fields(content[0])
	values, err := mapCPUStat(CPUStat)
	if err != nil {
		return map[string]int64{}, err
	}
	return values, nil
}

func mapCPUStat(utilData []string) (map[string]int64, error) {
	cpuStat := map[string]int64{}
	entries := []string{"user", "nice", "system", "idle", "iowait", "irq", "softirq", "steal", "guest", "guest_nice"}

	for i, entry := range entries {
		val, err := strconv.ParseInt(utilData[i+1], 10, 64)
		if err != nil {
			return nil, err

		}
		cpuStat[entry] = val
	}
	return cpuStat, nil
}
