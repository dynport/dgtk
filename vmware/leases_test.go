package vmware

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseLeases(t *testing.T) {
	Convey("ParseLeases", t, func() {
		leases, e := ParseLeases("fixtures/vmnet-dhcpd-vmnet8.leases")
		So(e, ShouldBeNil)
		So(len(leases), ShouldEqual, 8)
		So(leases[0].HardwareEthernet, ShouldEqual, "00:0c:29:80:9a:df")
		So(leases[0].Starts.Format("2006-01-02T15:04"), ShouldEqual, "2013-12-28T17:02")
		So(leases[0].Ends.Format("2006-01-02T15:04"), ShouldEqual, "2013-12-28T17:32")
	})
}
