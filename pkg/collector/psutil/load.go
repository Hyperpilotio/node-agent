package psutil

import (
	"fmt"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/shirou/gopsutil/load"
)

func loadAvg(nss []snap.Namespace) ([]snap.Metric, error) {
	defer timeSpent(time.Now(), "loadAvg")
	load, err := load.Avg()
	if err != nil {
		return nil, err
	}

	results := make([]snap.Metric, len(nss))

	for i, ns := range nss {
		switch ns.Element(len(ns) - 1).Value {
		case "load1":
			results[i] = snap.Metric{
				Namespace: ns,
				Data:      load.Load1,
				Unit:      "Load/1M",
				Timestamp: time.Now(),
			}
		case "load5":
			results[i] = snap.Metric{
				Namespace: ns,
				Data:      load.Load5,
				Unit:      "Load/5M",
				Timestamp: time.Now(),
			}
		case "load15":
			results[i] = snap.Metric{
				Namespace: ns,
				Data:      load.Load15,
				Unit:      "Load/15M",
				Timestamp: time.Now(),
			}
		default:
			return nil, fmt.Errorf("Requested load statistic %s is not found", ns.Element(len(ns)-1).Value)
		}
	}

	return results, nil
}

func getLoadAvgMetricTypes() []snap.Metric {
	defer timeSpent(time.Now(), "getLoadAvgMetricTypes")
	t := []int{1, 5, 15}
	mts := make([]snap.Metric, len(t))
	for i, te := range t {
		mts[i] = snap.Metric{
			Namespace: snap.NewNamespace("intel", "psutil", "load", fmt.Sprintf("load%d", te)),
			Unit:      fmt.Sprintf("Load/%dM", te),
		}
	}
	return mts
}
