package use

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	snap "github.com/hyperpilotio/node-agent/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUsePlugin(t *testing.T) {
	Convey("Create Use Collector", t, func() {
		useCol, _ := New()
		Convey("So NewUseCollector should not be nil", func() {
			So(useCol, ShouldNotBeNil)
		})
		Convey("So NewUseCollector should be of Use type", func() {
			So(useCol, ShouldHaveSameTypeAs, &Use{})
		})
	})

	Convey("Get Metrics ", t, func() {
		useCol, _ := New()
		var cfg = snap.Config{}

		Convey("So should return 26 types of metrics", func() {
			metrics, err := useCol.GetMetricTypes(cfg)
			So(len(metrics), ShouldBeGreaterThan, 13)
			So(err, ShouldBeNil)
		})
		Convey("So should check namespace", func() {
			metrics, err := useCol.GetMetricTypes(cfg)
			vcpuNamespace := metrics[0].Namespace.String()
			vcpu := regexp.MustCompile(`^/intel/use/compute/utilization`)
			So(true, ShouldEqual, vcpu.MatchString(vcpuNamespace))
			So(err, ShouldBeNil)

			vcpuNamespace1 := metrics[1].Namespace.String()
			vcpu1 := regexp.MustCompile(`^/intel/use/compute/saturation`)
			So(true, ShouldEqual, vcpu1.MatchString(vcpuNamespace1))
			So(err, ShouldBeNil)
		})

	})
	Convey("Collect Metrics", t, func() {
		useCol := &Use{}

		cfgNode := snap.NewConfig()

		pwd, err := os.Getwd()
		procPath = filepath.Join(pwd, "proc")
		diskStatPath = filepath.Join(procPath, "diskstats")
		cpuStatPath = filepath.Join(procPath, "stat")
		loadAvgPath = filepath.Join(procPath, "loadavg")
		memInfoPath = filepath.Join(procPath, "meminfo")
		vmStatPath = filepath.Join(procPath, "vmstat")
		So(err, ShouldBeNil)
		Convey("So should get memory saturation metrics", func() {
			metrics := []snap.Metric{{
				Namespace: snap.NewNamespace("intel", "use", "memory", "saturation"),
				Config:    cfgNode,
			}}
			collect, err := useCol.CollectMetrics(metrics)
			So(err, ShouldBeNil)
			So(collect[0].Data, ShouldNotBeNil)
			var expectedType float64
			So(collect[0].Data, ShouldHaveSameTypeAs, expectedType)
			So(len(collect), ShouldResemble, 1)
		})
		Convey("So should get memory utilization metrics", func() {
			metrics := []snap.Metric{{
				Namespace: snap.NewNamespace("intel", "use", "memory", "utilization"),
				Config:    cfgNode,
			}}
			collect, err := useCol.CollectMetrics(metrics)
			So(err, ShouldBeNil)
			So(collect[0].Data, ShouldNotBeNil)
			var expectedType float64
			So(collect[0].Data, ShouldHaveSameTypeAs, expectedType)
			So(len(collect), ShouldResemble, 1)
		})
		Convey("So should get compute utilization metrics", func() {
			metrics := []snap.Metric{{
				Namespace: snap.NewNamespace("intel", "use", "compute", "utilization"),
				Config:    cfgNode,
			}}
			collect, err := useCol.CollectMetrics(metrics)
			So(err, ShouldBeNil)
			So(collect[0].Data, ShouldNotBeNil)
			So(len(collect), ShouldResemble, 1)
		})
		Convey("So should get compute saturation metrics", func() {
			metrics := []snap.Metric{{
				Namespace: snap.NewNamespace("intel", "use", "compute", "saturation"),
				Config:    cfgNode,
			}}
			collect, err := useCol.CollectMetrics(metrics)
			So(err, ShouldBeNil)
			So(collect[0].Data, ShouldNotBeNil)
			var expectedType float64
			So(collect[0].Data, ShouldHaveSameTypeAs, expectedType)
			So(len(collect), ShouldResemble, 1)
		})
		Convey("So should get disk utilization metrics", func() {
			metrics := []snap.Metric{{
				Namespace: snap.NewNamespace("intel", "use", "storage", "sda", "utilization"),
				Config:    cfgNode,
			}}
			collect, err := useCol.CollectMetrics(metrics)
			So(err, ShouldBeNil)
			So(collect[0].Data, ShouldNotBeNil)
			So(len(collect), ShouldResemble, 1)
		})
		Convey("So should get disk saturation metrics", func() {
			metrics := []snap.Metric{{
				Namespace: snap.NewNamespace("intel", "use", "storage", "sda", "saturation"),
				Config:    cfgNode,
			}}
			collect, err := useCol.CollectMetrics(metrics)
			So(err, ShouldBeNil)
			So(collect[0].Data, ShouldNotBeNil)
			var expectedType float64
			So(collect[0].Data, ShouldHaveSameTypeAs, expectedType)
			So(len(collect), ShouldResemble, 1)
		})
	})
}
