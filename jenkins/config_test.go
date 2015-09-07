package jenkins

import (
	"encoding/xml"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSerializeConfig(t *testing.T) {
	Convey("Serialize Config", t, func() {
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

		Convey("Serialize Config", func() {
			config := &Config{}
			config.ShellBuilders = []*ShellBuilder{
				{Command: "whoami"},
			}
			b, e := xml.MarshalIndent(config, "", "  ")
			So(e, ShouldBeNil)
			So(b, ShouldNotBeNil)

			t.Logf("%s", string(b))
		})
	},
	)
}
