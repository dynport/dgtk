package c3

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
)

func New() *Chart {
	return &Chart{Data: &Data{}}
}

func NewTimeSeries(f string) *Chart {
	n := 0
	return &Chart{
		Data: &Data{X: "x", XFormat: f},
		Axis: map[string]*Axis{
			"x": &Axis{Type: "timeseries", Tick: &Tick{Format: "%Y-%m-%d %H%:%M:%S", Count: 5}},
			"y": &Axis{Min: &n}, // use pointer to 0
		},
	}
}

type Chart struct {
	Width   int              `json:"-"`
	Height  int              `json:"-"`
	BindTo  string           `json:"bindto,omitempty"`
	Data    *Data            `json:"data,omitempty"`
	Donut   *Donut           `json:"donut,omitempty"`
	Axis    map[string]*Axis `json:"axis,omitempty"`
	Tooltip interface{}      `json:"tooltip,omitempty"`
	Bar     *Bar             `json:"bar,omitempty"`
}

func (cp *Chart) Push(k string, v interface{}) {
	if cp.Data == nil {
		cp.Data = &Data{}
	}
	for i, c := range cp.Data.Columns {
		if len(c) > 0 && c[0] == k {
			cp.Data.Columns[i] = append(cp.Data.Columns[i], v)
			return
		}
	}
	cp.Data.Columns = append(cp.Data.Columns, []interface{}{k, v})
}

func (c *Chart) CSS() template.CSS {
	h, w := c.Height, c.Width
	if h == 0 {
		h = 150
	}
	if w == 0 {
		w = 300
	}
	return template.CSS(fmt.Sprintf("width:%dpx;height:%dpx", w, h))
}

func (c *Chart) ID() (string, error) {
	if c.BindTo == "" {
		return "", errors.New("BindTo must be set")
	}
	return strings.TrimPrefix(c.BindTo, "#"), nil
}

func (c *Chart) JS() (template.JS, error) {
	js, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return template.JS(js), nil
}
