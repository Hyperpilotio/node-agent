package psutil

import (
	"strings"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/shirou/gopsutil/disk"
)

func getPSUtilDiskUsage(path string) (*disk.UsageStat, error) {
	defer timeSpent(time.Now(), "getPSUtilDiskUsage")
	disk_usage, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}
	return disk_usage, nil
}

func getDiskUsageMetrics(nss []snap.Namespace, mounts []string) ([]snap.Metric, error) {
	defer timeSpent(time.Now(), "getDiskUsageMetrics")
	t := time.Now()
	var paths []disk.PartitionStat
	metrics := []snap.Metric{}
	namespaces := map[string][]string{}
	requested := map[string]snap.Namespace{}
	for _, ns := range nss {
		namespaces[ns.Strings()[len(ns.Strings())-1]] = ns.Strings()
		requested[ns.Strings()[len(ns.Strings())-1]] = ns
	}
	if strings.Contains(mounts[0], "physical") {
		parts, err := disk.Partitions(false)
		if err != nil {
			return nil, err
		}
		paths = parts
	} else if strings.Contains(mounts[0], "all") {
		parts, err := disk.Partitions(true)
		if err != nil {
			return nil, err
		}
		paths = parts
	} else {
		parts, err := disk.Partitions(true)
		if err != nil {
			return nil, err
		}
		for _, part := range parts {
			for _, mtpoint := range mounts {
				if part.Mountpoint == mtpoint {
					paths = append(paths, part)
				}
			}
		}
	}

	for _, path := range paths {
		data, err := getPSUtilDiskUsage(path.Mountpoint)
		if err != nil {
			return nil, err
		}
		tags := map[string]string{}
		tags["device"] = path.Device
		for _, namespace := range namespaces {
			if strings.Contains(strings.Join(namespace, "|"), "total") {
				nspace := make([]snap.NamespaceElement, len(requested["total"]))
				copy(nspace, requested["total"])
				nspace[3].Value = path.Mountpoint
				metrics = append(metrics, snap.Metric{
					Namespace: nspace,
					Data:      data.Total,
					Tags:      tags,
					Timestamp: t,
				})
			}
			if strings.Contains(strings.Join(namespace, "|"), "used") {
				nspace := make([]snap.NamespaceElement, len(requested["used"]))
				copy(nspace, requested["used"])
				nspace[3].Value = path.Mountpoint
				metrics = append(metrics, snap.Metric{
					Namespace: nspace,
					Data:      data.Used,
					Tags:      tags,
					Timestamp: t,
				})
			}
			if strings.Contains(strings.Join(namespace, "|"), "free") {
				nspace := make([]snap.NamespaceElement, len(requested["free"]))
				copy(nspace, requested["free"])
				nspace[3].Value = path.Mountpoint
				metrics = append(metrics, snap.Metric{
					Namespace: nspace,
					Data:      data.Free,
					Tags:      tags,
					Timestamp: t,
				})
			}
			if strings.Contains(strings.Join(namespace, "|"), "percent") {
				nspace := make([]snap.NamespaceElement, len(requested["percent"]))
				copy(nspace, requested["percent"])
				nspace[3].Value = path.Mountpoint
				metrics = append(metrics, snap.Metric{
					Namespace: nspace,
					Data:      data.UsedPercent,
					Tags:      tags,
					Timestamp: t,
				})
			}
		}
	}
	return metrics, nil
}

func getDiskUsageMetricTypes() []snap.Metric {
	defer timeSpent(time.Now(), "getDiskUsageMetricTypes")
	var mts []snap.Metric
	mts = append(mts, snap.Metric{
		Namespace: snap.NewNamespace("intel", "psutil", "disk").
			AddDynamicElement("mount_point", "Mount Point").
			AddStaticElement("total"),
	})
	mts = append(mts, snap.Metric{
		Namespace: snap.NewNamespace("intel", "psutil", "disk").
			AddDynamicElement("mount_point", "Mount Point").
			AddStaticElement("used"),
	})
	mts = append(mts, snap.Metric{
		Namespace: snap.NewNamespace("intel", "psutil", "disk").
			AddDynamicElement("mount_point", "Mount Point").
			AddStaticElement("free"),
	})
	mts = append(mts, snap.Metric{
		Namespace: snap.NewNamespace("intel", "psutil", "disk").
			AddDynamicElement("mount_point", "Mount Point").
			AddStaticElement("percent"),
	})
	return mts
}
