package stats

import "time"

func NewProgress(total int) *progress {
	return &progress{
		total: total, started: time.Now(),
	}
}

type progress struct {
	total   int
	count   int
	started time.Time
}

func (p *progress) Inc() {
	p.count++
}

func (s *progress) Count() int {
	return s.count
}

func (s *progress) Total() int {
	return s.total
}

func (s *progress) Perc() float64 {
	return 100.0 * float64(s.Count()) / float64(s.Total())
}

func (s *progress) ToGo() time.Duration {
	pending := float64(s.total - s.count)
	return time.Duration(pending/s.PerSecond()) * time.Second
}

func (s *progress) PerSecond() float64 {
	diff := time.Since(s.started)
	return float64(s.count) / diff.Seconds()
}
