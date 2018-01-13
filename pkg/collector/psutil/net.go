package psutil

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	psutilnet "github.com/shirou/gopsutil/net"
)

var netIOCounterLabels = map[string]label{
	"bytes_sent": label{
		unit:        "",
		description: "",
	},
	"bytes_recv": label{
		unit:        "",
		description: "",
	},
	"packets_sent": label{
		unit:        "",
		description: "",
	},
	"packets_recv": label{
		unit:        "",
		description: "",
	},
	"errin": label{
		unit:        "",
		description: "",
	},
	"errout": label{
		unit:        "",
		description: "",
	},
	"dropin": label{
		unit:        "",
		description: "",
	},
	"dropout": label{
		unit:        "",
		description: "",
	},
}

func netIOCounters(nss []snap.Namespace) ([]snap.Metric, error) {
	defer timeSpent(time.Now(), "netIOCounters")
	// gather accumulated metrics for all interfaces
	netsAll, err := psutilnet.IOCounters(false)
	if err != nil {
		return nil, err
	}

	// gather metrics per nic
	netsNic, err := psutilnet.IOCounters(true)
	if err != nil {
		return nil, err
	}

	results := []snap.Metric{}

	for _, ns := range nss {
		// set requested metric name from last namespace element
		metricName := ns.Element(len(ns) - 1).Value
		// check if requested metric is dynamic (requesting metrics for all nics)
		if ns[3].Value == "*" {
			for _, nic := range netsNic {
				// prepare namespace copy to update value
				// this will allow to keep namespace as dynamic (name != "")
				dyn := make([]snap.NamespaceElement, len(ns))
				copy(dyn, ns)
				dyn[3].Value = nic.Name
				// get requested metric value
				val, err := getNetIOCounterValue(&nic, metricName)
				if err != nil {
					return nil, err
				}
				tags, err := getInterfaceConfiguration(nic.Name)
				if err != nil {
					return nil, err
				}

				metric := snap.Metric{
					Namespace: dyn,
					Data:      val,
					Timestamp: time.Now(),
					Tags:      tags,
					Unit:      netIOCounterLabels[metricName].unit,
				}
				results = append(results, metric)
			}
		} else {
			stats := append(netsAll, netsNic...)

			// find stats for interface name or all nics
			stat := findNetIOStats(stats, ns[3].Value)
			if stat == nil {
				return nil, fmt.Errorf("Requested interface %s not found", ns[3].Value)
			}
			// get value for requested metric
			val, err := getNetIOCounterValue(stat, metricName)
			if err != nil {
				return nil, err
			}

			var tags map[string]string

			//for "all" interface there is no configuration
			if ns[3].Value != "all" {
				tags, err = getInterfaceConfiguration(ns[3].Value)
				if err != nil {
					return nil, err
				}
			}

			metric := snap.Metric{
				Namespace: ns,
				Data:      val,
				Tags:      tags,
				Timestamp: time.Now(),
				Unit:      netIOCounterLabels[metricName].unit,
			}
			results = append(results, metric)
		}
	}

	return results, nil
}

func findNetIOStats(nets []psutilnet.IOCountersStat, name string) *psutilnet.IOCountersStat {
	for _, net := range nets {
		if net.Name == name {
			return &net
		}
	}
	return nil
}

func getNetIOCounterValue(stat *psutilnet.IOCountersStat, name string) (uint64, error) {
	switch name {
	case "bytes_sent":
		return stat.BytesSent, nil
	case "bytes_recv":
		return stat.BytesRecv, nil
	case "packets_sent":
		return stat.PacketsSent, nil
	case "packets_recv":
		return stat.PacketsRecv, nil
	case "errin":
		return stat.Errin, nil
	case "errout":
		return stat.Errout, nil
	case "dropin":
		return stat.Dropin, nil
	case "dropout":
		return stat.Dropout, nil
	default:
		return 0, fmt.Errorf("Requested NetIOCounter statistic %s is not available", name)
	}
}

func getNetIOCounterMetricTypes() ([]snap.Metric, error) {
	defer timeSpent(time.Now(), "getNetIOCounterMetricTypes")
	mts := make([]snap.Metric, 0)

	for name, label := range netIOCounterLabels {
		//metrics which are the sum for all available nics
		mts = append(mts, snap.Metric{
			Namespace:   snap.NewNamespace("intel", "psutil", "net", "all", name),
			Description: label.description,
			Unit:        label.unit,
		})
		//dynamic metrics representing any nic
		mts = append(mts, snap.Metric{
			Namespace: snap.NewNamespace("intel", "psutil", "net").
				AddDynamicElement("interface_name", "network interface name").AddStaticElement(name),
			Description: label.description,
			Unit:        label.unit,
		})
	}

	return mts, nil
}

func getInterfaceConfiguration(ifaceName string) (map[string]string, error) {
	interfaceConfig, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, err
	}
	tags := make(map[string]string)
	hardwareAddr := interfaceConfig.HardwareAddr.String()
	if hardwareAddr != "" {
		tags["hardware_addr"] = hardwareAddr
	}
	if interfaceConfig.MTU != 0 {
		tags["mtu"] = strconv.Itoa(interfaceConfig.MTU)
	}
	return tags, nil
}
