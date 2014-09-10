package vmware

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVm(t *testing.T) {
	Convey("Modify", t, func() {
		Convey("With memsize existing", func() {
			e := os.RemoveAll("tmp")
			So(e, ShouldBeNil)
			e = os.MkdirAll("tmp", 0755)
			So(e, ShouldBeNil)
			e = ioutil.WriteFile("tmp/test.vmx", []byte(vmx), 0644)
			So(e, ShouldBeNil)
			vm := &Vm{Path: "tmp/test.vmx"}
			e = vm.ModifyMemory(2048)
			So(e, ShouldBeNil)
			b, e := ioutil.ReadFile("tmp/test.vmx")
			So(e, ShouldBeNil)
			So(b, ShouldNotBeNil)
			So(string(b), ShouldContainSubstring, `memsize = "2048"`+"\n")
		})

		Convey("With memsize not existing", func() {
			e := os.RemoveAll("tmp")
			So(e, ShouldBeNil)
			e = os.MkdirAll("tmp", 0755)
			So(e, ShouldBeNil)
			e = ioutil.WriteFile("tmp/test.vmx", []byte(""), 0644)
			So(e, ShouldBeNil)
			vm := &Vm{Path: "tmp/test.vmx"}
			e = vm.ModifyMemory(2048)
			So(e, ShouldBeNil)
			b, e := ioutil.ReadFile("tmp/test.vmx")
			So(e, ShouldBeNil)
			So(b, ShouldNotBeNil)
			So(string(b), ShouldContainSubstring, `memsize = "2048"`)
		})

		Convey("Modify cpu", func() {
			e := os.RemoveAll("tmp")
			So(e, ShouldBeNil)
			e = os.MkdirAll("tmp", 0755)
			So(e, ShouldBeNil)
			e = ioutil.WriteFile("tmp/test.vmx", []byte(""), 0644)
			So(e, ShouldBeNil)
			vm := &Vm{Path: "tmp/test.vmx"}
			e = vm.ModifyCpu(4)
			So(e, ShouldBeNil)
			b, e := ioutil.ReadFile("tmp/test.vmx")
			So(e, ShouldBeNil)
			So(b, ShouldNotBeNil)
			So(string(b), ShouldContainSubstring, `numvcpus = "4"`)
		})
	})
}

const vmx = `.encoding = "UTF-8"
config.version = "8"
virtualHW.version = "10"
vcpu.hotadd = "TRUE"
scsi0.present = "TRUE"
scsi0.virtualDev = "lsilogic"
sata0.present = "TRUE"
memsize = "1024"
mem.hotadd = "TRUE"
scsi0:0.present = "TRUE"`
