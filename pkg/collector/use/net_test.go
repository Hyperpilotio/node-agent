package use

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shirou/gopsutil/net"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNetUsePlugin(t *testing.T) {

	Convey("Should return 1 when interface dosn't exist", t, func() {

		linkspeed := getLinkSpeedforIface("mock0")
		So(linkspeed, ShouldResemble, 1.0)

	})
	Convey("Should return return proper linkspeed when interface exist", t, func() {

		pwd, err := os.Getwd()
		So(err, ShouldBeNil)
		sysFsNetPath = filepath.Join(pwd, "sys", "class", "net")
		linkspeed := getLinkSpeedforIface("eth0")
		So(linkspeed, ShouldResemble, 125000.0)

	})
	Convey("Should return an error when interface doesn't exist", t, func() {

		stats, err := getNicStatistic("mock0")
		So(stats, ShouldResemble, net.IOCountersStat{Name: "", BytesSent: 0x0, BytesRecv: 0x0, PacketsSent: 0x0, PacketsRecv: 0x0, Errin: 0x0, Errout: 0x0, Dropin: 0x0, Dropout: 0x0, Fifoin: 0x0, Fifoout: 0x0})
		So(err, ShouldNotBeNil)

	})

}
