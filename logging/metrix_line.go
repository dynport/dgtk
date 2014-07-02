package logging

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type MetrixLine struct {
	Host      string
	Metric    string
	Timestamp time.Time
	Value     int64
	Tags      MetrixTags
}

type MetrixTags map[string]string

func (m *MetrixLine) Parse(line string) error {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return fmt.Errorf("expected at least 6 fields")
	}
	m.Host = fields[1]
	i, e := strconv.ParseInt(fields[4], 10, 64)
	if e != nil {
		return e
	}
	m.Timestamp = time.Unix(i, 0)
	m.Metric = fields[3]
	m.Value, e = strconv.ParseInt(fields[5], 10, 64)
	if e != nil {
		return e
	}
	if len(fields) > 6 {
		m.Tags = MetrixTags{}
		for _, f := range fields[6:] {
			parts := strings.SplitN(f, "=", 2)
			if len(parts) == 2 {
				m.Tags[parts[0]] = parts[1]

			}

		}
	}
	return nil
}
