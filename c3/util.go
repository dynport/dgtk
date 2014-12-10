package c3

const Spline = "spline"

type Ratio struct {
	Ration float64 `json:"ration,omitempty"`
}

// columns: [
//   ['x', '2013-01-01', '2013-01-02', '2013-01-03', '2013-01-04', '2013-01-05', '2013-01-06'],
//   // ['x', '20130101', '20130102', '20130103', '20130104', '20130105', '20130106'],
//   ['data1', 30, 200, 100, 400, 150, 250],
//   ['data2', 130, 340, 200, 500, 250, 350]
// ]
const TimeSeries = "timeseries"

type Tick struct {
	Format string `json:"format,omitempty"` // %Y-%m-%d
	Count  int    `json:"count,omitempty"`
}
