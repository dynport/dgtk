package vmware

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestVmx(t *testing.T) {
	Convey("Vmx", t, func() {
		vmx := &Vmx{}
		e := vmx.Parse("fixtures/ubuntu.13.04.vmx")
		So(e, ShouldBeNil)
		So(vmx.MacAddress, ShouldEqual, "00:0c:29:79:29:aa")
		So(vmx.CleanShutdown, ShouldBeTrue)
		So(vmx.SoftPowerOff, ShouldBeTrue)
	})
}
