package logging

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricLine(t *testing.T) {
	Convey("Parse", t, func() {
		m := &MetrixLine{}
		line := `2014-06-03T08:03:01.357642+00:00 krusty-he-284430 metrix.notice[14506]: processes.Vsize 1401782581 10 pid=320 ppid=2 comm=(md1_raid1) name=md1_raid1 state=S host=krusty-he-284430`
		e := m.Parse(line)
		So(e, ShouldBeNil)

		So(m.Host, ShouldEqual, "krusty-he-284430")
		So(m.Metric, ShouldEqual, "processes.Vsize")
		So(m.Timestamp.Format("2006-01-02T15:04"), ShouldEqual, "2014-06-03T10:03")
		So(m.Value, ShouldEqual, 10)

		So(len(m.Tags), ShouldEqual, 6)
		So(m.Tags["host"], ShouldEqual, "krusty-he-284430")
	})
}
