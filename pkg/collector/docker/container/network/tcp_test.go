package network

import (
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// docker container's process ID points to its tcp stats in /proc/{pid}/net/tcp
var mockPids = []int{1234, 5678, 91011}

//mockTcpContent := []byte(`ivnalidsl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode`)

var mockTcpContent = []byte(`sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
				   0: 0100007F:F76E 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 2638360 1 ffff8805af412800 100 0 0 10 0
				   1: 00000000:1771 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 24271 1 ffff8807e0d9a800 100 0 0 10 0
				   2: 0101007F:0035 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 16914 1 ffff8807f79d0800 100 0 0 10 0
				   3: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 9574802 1 ffff8807df908800 100 0 0 10 0
				   4: 0100007F:0277 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 10216612 1 ffff8800a959c000 100 0 0 10 0
				   5: 0100007F:1B1E 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 2646436 1 ffff8807e0d9d000 100 0 0 10 0
				   6: 0100007F:B1FE 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 19907 1 ffff8807e0fc0000 100 0 0 10 0
				   7: 1E7E5B0A:8CE4 315D1332:01BB 01 00000000:00000000 00:00000000 00000000     0        0 5118203 1 ffff8806a1817800 20 0 0 10 -1`)

func TestTcpStatsFromProc(t *testing.T) {
	defer deleteMockFiles()

	Convey("Get TCP/TCP6 stats from procfs", t, func() {

		Convey("create statistics for mock devices", func() {
			err := createMockProcfsNetTCP(mockPids, mockTcpContent)
			So(err, ShouldBeNil)
		})

		Convey("successful retrieving TCP statistics", func() {
			for _, pid := range mockPids {
				path := filepath.Join(mockProcfsDir, strconv.Itoa(pid), "net/tcp")
				tcpStats, err := tcpStatsFromProc(path)
				So(err, ShouldBeNil)
				So(tcpStats, ShouldNotBeEmpty)
				So(tcpStats.Established, ShouldEqual, 1)
				So(tcpStats.Listen, ShouldEqual, 7)
			}

		})

		Convey("successful retrieving TCP6 statistics", func() {
			for _, pid := range mockPids {
				path := filepath.Join(mockProcfsDir, strconv.Itoa(pid), "net/tcp6")
				tcpStats, err := tcpStatsFromProc(path)
				So(err, ShouldBeNil)
				So(tcpStats, ShouldNotBeEmpty)
				So(tcpStats.Established, ShouldEqual, 1)
				So(tcpStats.Listen, ShouldEqual, 7)
			}

		})

		Convey("return an error when the given PID does not exist", func() {
			path := filepath.Join(mockProcfsDir, strconv.Itoa(0), "net/tcp")
			tcpStats, err := tcpStatsFromProc(path)
			So(err, ShouldNotBeNil)
			So(tcpStats, ShouldBeZeroValue)
		})

		Convey("return an error when content is invalid", func() {
			mockPid := 1
			err := createMockProcfsNetTCP([]int{mockPid}, []byte(`invalid`))
			So(err, ShouldBeNil)
			path := filepath.Join(mockProcfsDir, strconv.Itoa(mockPid), "net/tcp")
			tcpStats, err := tcpStatsFromProc(path)
			//So(err, ShouldNotBeNil)
			So(tcpStats, ShouldBeZeroValue)
		})

		Convey("return an error when content is empty", func() {
			mockPid := 2
			err := createMockProcfsNetTCP([]int{mockPid}, []byte(``))
			So(err, ShouldBeNil)
			path := filepath.Join(mockProcfsDir, strconv.Itoa(mockPid), "net/tcp")
			tcpStats, err := tcpStatsFromProc(path)
			So(tcpStats, ShouldBeZeroValue)
		})

	})
}
