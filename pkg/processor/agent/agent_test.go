package agent

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessor(t *testing.T) {
	Convey("Test Processor", t, func() {
		Convey("Process int metric", func() {
			// metrics := []plugin.Metric{
			// 	plugin.Metric{
			// 		Namespace: plugin.NewNamespace("x", "y", "z"),
			// 		Config:    map[string]interface{}{"pw": "123aB"},
			// 		Data:      345678,
			// 		Tags:      map[string]string{"hello": "world"},
			// 		Unit:      "int",
			// 		Timestamp: time.Now(),
			// 	},
			// }
			// mts, err := p.Process(metrics, plugin.Config{})
			// So(mts, ShouldNotBeNil)
			// So(err, ShouldBeNil)
			// So(mts[0].Data, ShouldEqual, 876543)
		})

		Convey("Test GetConfigPolicy", func() {
			p := GodddQoSProcessor{}
			_, err := p.GetConfigPolicy()

			Convey("No error returned", func() {
				So(err, ShouldBeNil)
			})

		})
	})
}
