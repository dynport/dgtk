package progress

import (
	"fmt"
	"runtime"
	"time"
)

type Status struct {
	Started  time.Time
	Now      time.Time
	Current  int
	MemStats runtime.MemStats
	Total    *int
}

func (status *Status) String() string {
	s := fmt.Sprintf("total_time=%.06f per_second=%.01f", status.RunningSince().Seconds(), status.PerSecond())
	if status.Total != nil {
		l := IntLen(*status.Total)
		s = fmt.Sprintf("cnt=%0*d/%0*d ", l, status.Current, l, *status.Total) + s + fmt.Sprintf(" eta=%.01f", status.ETA().Seconds())
	} else {
		s = fmt.Sprintf("cnt=%d ", status.Current) + s
	}
	if status.MemStats.Alloc > 0 {
		s += fmt.Sprintf(" mem_alloc=%s", SizePretty(float64(status.MemStats.Alloc)))
	}
	return s
}

func (s *Status) PerSecond() float64 {
	return float64(s.Current) / s.RunningSince().Seconds()
}

func (s *Status) ETA() *time.Duration {
	if s.Total == nil {
		return nil
	}
	d := time.Duration(float64(*s.Total-s.Current)/s.PerSecond()) * time.Second
	return &d
}

func (s *Status) RunningSince() time.Duration {
	return s.Now.Sub(s.Started)
}
