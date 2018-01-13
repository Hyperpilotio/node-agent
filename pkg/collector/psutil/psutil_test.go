package psutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
		// 	Convey("So config policy should be a cpolicy.ConfigPolicy", func() {
		// 		So(configPolicy, ShouldHaveSameTypeAs, plugin.ConfigPolicy{})
		// 	})
		// })
	})
}
