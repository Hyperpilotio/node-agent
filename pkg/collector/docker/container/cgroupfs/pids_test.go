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

type PidsSuite struct {
	suite.Suite
	pidsPath string
}

func (suite *PidsSuite) SetupSuite() {
	suite.pidsPath = "/tmp/pids_test"
	err := os.Mkdir(suite.pidsPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.writeFile(filepath.Join(suite.pidsPath, "pids.current"), []byte("1"))
	suite.writeFile(filepath.Join(suite.pidsPath, "pids.max"), []byte("2"))

}

func (suite *PidsSuite) TearDownSuite() {
	err := os.RemoveAll(suite.pidsPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestPidsSuite(t *testing.T) {
	suite.Run(t, &PidsSuite{})
}

func (suite *PidsSuite) TestPidsGetStats() {
	Convey("collecting data from cpuset controller", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.pidsPath}
		pids := Pids{}
		err := pids.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.PidsStats.Current, ShouldEqual, 1)
		So(stats.Cgroups.PidsStats.Limit, ShouldEqual, 2)
	})
}

func (suite *PidsSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
