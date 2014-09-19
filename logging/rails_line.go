package logging

import (
	"fmt"
	"strings"
	"time"
)

type RailsLine struct {
	Uuid             string  `json:"uuid,omitempty"`
	Action           string  `json:"action,omitempty"`
	Type             string  `json:"type,omitempty"`
	ExecutionTime    float64 `json:"execution_time,omitempty"`
	TotalTime        float64 `json:"total_time,omitempty"`
	ViewTime         float64 `json:"view_time,omitempty"`
	ActiveRecordTime float64 `json:"active_record_time,omitempty"`
}

func (r *RailsLine) Parse(raw string) error {
	fields := strings.Fields(raw)
	last := ""
	state := 0
	var e error
	for _, f := range fields {
		if strings.HasPrefix(f, "[") && strings.HasSuffix(f, "]") && len(f) == 38 {
			r.Uuid = strings.TrimSuffix(strings.TrimPrefix(f, "["), "]")
		} else if strings.HasPrefix(f, "(") && strings.HasSuffix(f, "ms)") {
			r.ExecutionTime, e = parseAndNormalizeDuration(f)
			if e != nil {
				return e
			}
		} else if r.Uuid != "" {
			switch {
			case last == "Views:":
				r.ViewTime, e = parseAndNormalizeDuration(f)
				if e != nil {
					return e
				}
			case last == "ActiveRecord:":
				r.ActiveRecordTime, e = parseAndNormalizeDuration(f)
				if e != nil {
					return e
				}
			case last == "in" && strings.HasSuffix(f, "ms"):
				r.TotalTime, e = parseAndNormalizeDuration(f)
				r.Type = "completed"
				if e != nil {
					return e
				}
			case f == "Processing":
				state = stateProcessing
			case f == "by" && state == stateProcessing:
				state = stateProcessingBy
			case state == stateProcessingBy:
				r.Action = f
				r.Type = "processing"
				state = 0
			default:
				state = 0
			}
		}
		last = f
	}
	if r.Uuid == "" {
		return fmt.Errorf("expected to find uuid in log but did not")
	}
	return nil
}

const (
	stateProcessing = iota + 1
	stateProcessingBy
)

func parseAndNormalizeDuration(s string) (float64, error) {
	d, e := time.ParseDuration(strings.TrimSuffix(strings.TrimPrefix(s, "("), ")"))
	if e != nil {
		return 0, e
	}
	return d.Seconds(), nil

}
