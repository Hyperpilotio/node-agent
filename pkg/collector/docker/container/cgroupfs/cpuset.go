package cgroupfs

import (
	"path/filepath"

	"github.com/hyperpilotio/node-agent/pkg/collector/docker/container"
)

// CpuSet implements StatGetter interface
type CpuSet struct{}

// GetStats reads cpuset metrics from Cpuset Group
func (cs *CpuSet) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	cpus, err := parseStrValue(filepath.Join(path, "cpuset.cpus"))
	if err != nil {
		return err
	}

	mems, err := parseStrValue(filepath.Join(path, "cpuset.mems"))
	if err != nil {
		return err
	}

	memmig, err := parseIntValue(filepath.Join(path, "cpuset.memory_migrate"))
	if err != nil {
		return err
	}

	cpuexc, err := parseIntValue(filepath.Join(path, "cpuset.cpu_exclusive"))
	if err != nil {
		return err
	}

	memexc, err := parseIntValue(filepath.Join(path, "cpuset.mem_exclusive"))
	if err != nil {
		return err
	}

	stats.Cgroups.CpuSetStats.Cpus = cpus
	stats.Cgroups.CpuSetStats.Mems = mems
	stats.Cgroups.CpuSetStats.MemoryMigrate = memmig
	stats.Cgroups.CpuSetStats.CpuExclusive = cpuexc
	stats.Cgroups.CpuSetStats.MemoryExclusive = memexc

	return nil
}
