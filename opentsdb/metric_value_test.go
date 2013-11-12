package opentsdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var values = []struct{ Line, Expected string }{
	{Line: "load.Load1m 1383516272 18 host=cnc-3b06098c", Expected: "cnc-3b06098c"},
	{Line: "load.Load1m 1383516272 18", Expected: ""},
}

func TestExtractTag(t *testing.T) {
	for _, line := range values {
		m := &MetricValue{}
		e := m.Parse(line.Line)
		if e != nil {
			t.Errorf("parsing %s returned error %s", line, e.Error())
		} else {
			result := m.ExtractTag("host")
			assert.Equal(t, line.Expected, result, "expected host to be %q but was %q", line.Expected, result)
		}
	}
}
