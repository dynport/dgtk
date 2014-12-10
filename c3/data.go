package c3

type AxisType string

// spline, area, area-spline, bar, pie, donut, gauge (percentages)
type Type string //

type Data struct {
	Columns [][]interface{} `json:"columns,omitempty"`
	X       string          `json:"x,omitempty"`
	XFormat string          `json:"xFormat,omitempty"`
	Type    Type            `json:"type,omitempty"` // spline, bar, pie
	Types   map[string]Type `json:"types,omitempty"`
	Groups  [][]string      `json:"groups,omitempty"` // which charts to stack, e.g. [['data1', 'data2']]
}
