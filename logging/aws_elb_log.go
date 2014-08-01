package logging

import (
	"strconv"
	"strings"
	"time"
)

const (
	elbLogTimestamp = iota
	elbLogElb
	elbLogClientAndPort
	elbLogBackendAndPort
	elbLogRequestProcessingTime
	elbLogBackendProcessingTime
	elbLogResponseProcessingTime
	elbLogElbStatusCode
	elbLogBackendStatusCode
	elbLogReceivedBytes
	elbLogSentBytes
	elbLogRequest
)

type ElasticLoadBalancerLog struct {
	Timestamp              time.Time
	Elb                    string
	ClientAndPort          string
	BackendAndPort         string
	RequestProcessingTime  float64
	BackendProcessingTime  float64
	ResponseProcessingTime float64
	ElbStatusCode          int
	BackendStatusCode      int
	ReceivedBytes          int
	SentBytes              int
	Method                 string
	Url                    string
	Action                 string
}

func (l *ElasticLoadBalancerLog) Load(raw string) error {
	var e error
	fields := Fields(raw)
	for i, f := range fields {
		switch i {
		case elbLogTimestamp:
			l.Timestamp, e = time.Parse("2006-01-02T15:04:05.999999Z", f)
			if e != nil {
				return e
			}
		case elbLogElb:
			l.Elb = f
		case elbLogClientAndPort:
			l.ClientAndPort = f
		case elbLogBackendAndPort:
			l.BackendAndPort = f
		case elbLogRequestProcessingTime:
			l.RequestProcessingTime, e = strconv.ParseFloat(f, 64)
			if e != nil {
				return e
			}
		case elbLogBackendProcessingTime:
			l.BackendProcessingTime, e = strconv.ParseFloat(f, 64)
			if e != nil {
				return e
			}
		case elbLogResponseProcessingTime:
			l.ResponseProcessingTime, e = strconv.ParseFloat(f, 64)
			if e != nil {
				return e
			}
		case elbLogElbStatusCode:
			l.ElbStatusCode, e = strconv.Atoi(f)
			if e != nil {
				return e
			}
		case elbLogBackendStatusCode:
			l.BackendStatusCode, e = strconv.Atoi(f)
			if e != nil {
				return e
			}
		case elbLogReceivedBytes:
			l.ReceivedBytes, e = strconv.Atoi(f)
			if e != nil {
				return e
			}
		case elbLogSentBytes:
			l.SentBytes, e = strconv.Atoi(f)
			if e != nil {
				return e
			}
		case elbLogRequest:
			parts := strings.Split(f, " ")
			if len(parts) == 3 {
				l.Method = parts[0]
				l.Url = parts[1]
				parts := strings.Split(l.Url, "/")
				if len(parts) > 0 {
					l.Action = parts[len(parts)-1]
				}
			}
		}
	}
	return nil
}
