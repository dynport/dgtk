package c3

type Axis struct {
	Min  *int     `json:"min,omitempty"`
	Max  *int     `json:"max,omitempty"`
	Type AxisType `json:"type,omitempty"` // timeseries
	Tick *Tick    `json:"tick,omitempty"`
}
