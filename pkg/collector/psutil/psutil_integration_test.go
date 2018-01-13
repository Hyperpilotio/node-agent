package psutil

import (
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/hyperpilotio/node-agent/pkg/snap"
)

func TestPsutilGetMetricTypes(t *testing.T) {
	Convey("psutil collector", t, func() {
		p := NewPsutilCollector()
		Convey("get metric types", func() {
			config := snap.Config{}

			metric_types, err := p.GetMetricTypes(config)
			So(err, ShouldBeNil)
			So(metric_types, ShouldNotBeNil)
			So(metric_types, ShouldNotBeEmpty)
			//55 collectable metrics
			So(len(metric_types), ShouldEqual, 55)
		})
	})

}

func TestPsutilCollectMetrics(t *testing.T) {
	Convey("psutil collector", t, func() {
		p := NewPsutilCollector()
		Convey("collect metrics", func() {
			config := snap.Config{
				"mount_points": "/|/dev|/run",
			}

			mts := []snap.Metric{
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "load", "load1"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "load", "load5"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "load", "load15"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "total"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "available"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "used"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "used_percent"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "free"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "active"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "inactive"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "buffers"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "cached"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "vm", "wired"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "disk", "used"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "nice"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "system"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "iowait"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "guest"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "stolen"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "idle"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "guest_nice"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "irq"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "softirq"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "*", "steal"),
					Config:    config,
				},

				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "bytes_sent"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "bytes_recv"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "disk", "*", "total"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "disk", "*", "used"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "disk", "*", "free"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "disk", "*", "percent"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "disk", "*", "used"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "bytes_recv"),
					Config:    config,
				},

				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "packets_sent"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "packets_recv"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "errin"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "*", "errout"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "lo", "dropin"),
					Config:    config,
				},
				snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "net", "all", "dropout"),
					Config:    config,
				},
			}
			if runtime.GOOS != "darwin" {
				mts = append(mts, snap.Metric{
					Namespace: snap.NewNamespace("intel", "psutil", "cpu", "cpu0", "user"),
				})
			}
			metrics, err := p.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
		})
		Convey("get metric types", func() {
			mts, err := p.GetMetricTypes(snap.Config{})
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
		})

	})
}

func TestPSUtilPlugin(t *testing.T) {
	Convey("Create PSUtil Collector", t, func() {
		psCol := NewPsutilCollector()
		Convey("So psCol should not be nil", func() {
			So(psCol, ShouldNotBeNil)
		})
		Convey("So psCol should be of Psutil type", func() {
			So(psCol, ShouldHaveSameTypeAs, &Psutil{})
		})
		// Convey("psCol.GetConfigPolicy() should return a config policy", func() {
		// 	configPolicy, _ := psCol.GetConfigPolicy()
		// 	Convey("So config policy should not be nil", func() {
		// 		So(configPolicy, ShouldNotBeNil)
		// 	})
		// 	Convey("So config policy should be a snap.ConfigPolicy", func() {
		// 		So(configPolicy, ShouldHaveSameTypeAs, snap.ConfigPolicy{})
		// 	})
		// })
	})
}
