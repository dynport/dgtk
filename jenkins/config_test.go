package jenkins

import (
	"encoding/xml"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSerializeConfig(t *testing.T) {
	Convey("Serialize Config", t, func() {
		So(1, ShouldEqual, 1)
		Convey("Serialize SCM", func() {
			scm := &Scm{
				Class:  "hudson.plugins.git.GitSCM",
				Plugin: "git@2.0",
			}
			b, e := xml.Marshal(scm)
			So(e, ShouldBeNil)
			if e != nil {
				So(e.Error(), ShouldEqual, "")
			}
			So(string(b), ShouldContainSubstring, `class="hudson.plugins.git.GitSCM`)
		})
	})
}
