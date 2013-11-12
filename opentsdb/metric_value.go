package opentsdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Data type for a metric value.
type MetricValue struct {
	Key   string    // Metric's key.
	Value float64   // Metric's value.
	Time  time.Time // Timestamp of metric recording.
	Tags  string    // Tags the metric has set.
}

func (value *MetricValue) ExtractTag(name string) string {
	for _, tag := range strings.Fields(value.Tags) {
		if strings.HasPrefix(tag, name+"=") {
			return tag[len(name) + 1:]
		}
	}
	return ""
}

func (mv *MetricValue) Parse(line string) error {
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 3 {
		logger.Debug("failed to parse line:", line)
		return fmt.Errorf("failed to parse line")
	}
	mv.Key = parts[0]
	if len(parts) == 4 {
		mv.Tags = parts[3]
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		logger.Debug("failed to parse timestamp:", parts[1])
		return err
	}
	mv.Time = time.Unix(timestamp, 0)

	value, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		logger.Debug("failed to parse value:", parts[2])
		return err
	}
	mv.Value = value
	return nil
}

func (mv *MetricValue) String() string {
	return fmt.Sprintf("%s %s %.01f %s", mv.Time.Format("2006-01-02T15:04:05"), mv.Key, mv.Value, mv.Tags)
}
