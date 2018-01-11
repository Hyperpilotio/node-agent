package cgroupfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"

	"github.com/hyperpilotio/node-agent/pkg/collector/docker/container"
)

type CpuSetSuite struct {
	suite.Suite
	cpusetPath string
}

func (suite *CpuSetSuite) SetupSuite() {
	suite.cpusetPath = "/tmp/cpuset_test"
	err := os.Mkdir(suite.cpusetPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.memory_migrate"), []byte("1"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.cpu_exclusive"), []byte("2"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.mem_exclusive"), []byte("3"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.mems"), []byte("4"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.cpus"), []byte("5"))
}

func (suite *CpuSetSuite) TearDownSuite() {
	err := os.RemoveAll(suite.cpusetPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestCpuSetSuite(t *testing.T) {
	suite.Run(t, &CpuSetSuite{})
}

func (suite *CpuSetSuite) TestCpuGetStats() {
	Convey("collecting data from cpuset controller", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.cpusetPath}
		cpu := CpuSet{}
		err := cpu.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.CpuSetStats.MemoryMigrate, ShouldEqual, 1)
		So(stats.Cgroups.CpuSetStats.CpuExclusive, ShouldEqual, 2)
		So(stats.Cgroups.CpuSetStats.MemoryExclusive, ShouldEqual, 3)
		So(stats.Cgroups.CpuSetStats.Mems, ShouldEqual, "4")
		So(stats.Cgroups.CpuSetStats.Cpus, ShouldEqual, "5")
	})
}

func (suite *CpuSetSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
