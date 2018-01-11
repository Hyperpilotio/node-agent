package psutil

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/shirou/gopsutil/cpu"
)

var cpuLabels = map[string]label{
	"user": label{
		description: "",
		unit:        "",
	},
	"system": label{
		description: "",
		unit:        "",
	},
	"idle": label{
		description: "",
		unit:        "",
	},
	"nice": label{
		description: "",
		unit:        "",
	},
	"iowait": label{
		description: "",
		unit:        "",
	},
	"irq": label{
		description: "",
		unit:        "",
	},
	"softirq": label{
		description: "",
		unit:        "",
	},
	"steal": label{
		description: "",
		unit:        "",
	},
	"guest": label{
		description: "",
		unit:        "",
	},
	"guest_nice": label{
		description: "",
		unit:        "",
	},
	"stolen": label{
		description: "",
		unit:        "",
	},
}

func cpuTimes(nss []snap.Namespace) ([]snap.Metric, error) {
	// gather metrics per each cpu
	defer timeSpent(time.Now(), "cpuTimes")
	timesCPUs, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}

	// gather accumulated metrics for all cpus
	timesAll, err := cpu.Times(false)
	if err != nil {
		return nil, err
	}

	results := []snap.Metric{}

	for _, ns := range nss {
		// set requested metric name from last namespace element
		metricName := ns.Element(len(ns) - 1).Value
		// check if requested metric is dynamic (requesting metrics for all cpu ids)
		if ns[3].Value == "*" {
			for _, timesCPU := range timesCPUs {
				// prepare namespace copy to update value
				// this will allow to keep namespace as dynamic (name != "")
				dyn := make([]snap.NamespaceElement, len(ns))
				copy(dyn, ns)
				dyn[3].Value = timesCPU.CPU
				// get requested metric value
				val, err := getCPUTimeValue(&timesCPU, metricName)
				if err != nil {
					return nil, err
				}
				metric := snap.Metric{
					Namespace: dyn,
					Data:      val,
					Timestamp: time.Now(),
					Unit:      cpuLabels[metricName].unit,
				}
				results = append(results, metric)
			}
		} else {
			timeStats := append(timesAll, timesCPUs...)
			// find stats for interface name or all cpus
			timeStat := findCPUTimeStat(timeStats, ns[3].Value)
			if timeStat == nil {
				return nil, fmt.Errorf("Requested cpu id %s not found", ns[3].Value)
			}
			// get requested metric value from struct
			val, err := getCPUTimeValue(timeStat, metricName)
			if err != nil {
				return nil, err
			}
			metric := snap.Metric{
				Namespace: ns,
				Data:      val,
				Timestamp: time.Now(),
				Unit:      cpuLabels[metricName].unit,
			}
			results = append(results, metric)
		}
	}

	return results, nil
}

func findCPUTimeStat(timeStats []cpu.TimesStat, name string) *cpu.TimesStat {
	for _, timeStat := range timeStats {
		if timeStat.CPU == name {
			return &timeStat
		}
	}
	return nil
}

func getCPUTimeValue(stat *cpu.TimesStat, name string) (float64, error) {
	switch name {
	case "user":
		return stat.User, nil
	case "system":
		return stat.System, nil
	case "idle":
		return stat.Idle, nil
	case "nice":
		return stat.Nice, nil
	case "iowait":
		return stat.Iowait, nil
	case "irq":
		return stat.Irq, nil
	case "softirq":
		return stat.Softirq, nil
	case "steal":
		return stat.Steal, nil
	case "guest":
		return stat.Guest, nil
	case "guest_nice":
		return stat.GuestNice, nil
	case "stolen":
		return stat.Stolen, nil
	default:
		return 0, fmt.Errorf("Requested CPUTime statistic %s is not available", name)
	}
}

func getCPUTimesMetricTypes() ([]snap.Metric, error) {
	defer timeSpent(time.Now(), "getCPUTimesMetricTypes")
	//passing true to CPUTimes indicates per CPU
	mts := []snap.Metric{}
	switch runtime.GOOS {
	case "linux", "darwin":
		for k, label := range cpuLabels {
			mts = append(mts, snap.Metric{
				Namespace:   snap.NewNamespace("intel", "psutil", "cpu").AddDynamicElement("cpu_id", "physical cpu id").AddStaticElement(k),
				Description: label.description,
				Unit:        label.unit,
			})
			mts = append(mts, snap.Metric{
				Namespace:   snap.NewNamespace("intel", "psutil", "cpu", "cpu-total").AddStaticElement(k),
				Description: label.description,
				Unit:        label.unit,
			})
		}
	default:
		return nil, fmt.Errorf("%s not supported by plugin", runtime.GOOS)
	}
	return mts, nil
}
