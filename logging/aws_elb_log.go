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
	elbLogUserAgent
	elbLogSSLCipher
	elbLogSSLProtocol
)

type ElasticLoadBalancerLog struct {
	Timestamp              time.Time `json:"timestamp,omitempty"`
	Elb                    string    `json:"elb,omitempty"`
	ClientAndPort          string    `json:"client_and_port,omitempty"`
	BackendAndPort         string    `json:"backend_and_port,omitempty"`
	RequestProcessingTime  float64   `json:"request_processing_time,omitempty"`
	BackendProcessingTime  float64   `json:"backend_processing_time,omitempty"`
	ResponseProcessingTime float64   `json:"response_processing_time,omitempty"`
	ElbStatusCode          int       `json:"elb_status_code,omitempty"`
	BackendStatusCode      int       `json:"backend_status_code,omitempty"`
	ReceivedBytes          int       `json:"received_bytes,omitempty"`
	SentBytes              int       `json:"sent_bytes,omitempty"`
	Method                 string    `json:"method,omitempty"`
	Url                    string    `json:"url,omitempty"`
	UserAgent              string    `json:"user_agent,omitempty"`
	SSLCipher              string    `json:"ssl_cipher,omitempty"`
	SSLProtocol            string    `json:"ssl_protocol,omitempty"`
	RAW                    string    `json:"raw,omitempty"`
}

func (l *ElasticLoadBalancerLog) Load(raw string) error {
	l.RAW = raw
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
			}
		case elbLogUserAgent:
			l.UserAgent = f
		case elbLogSSLCipher:
			if f != "-" {
				l.SSLCipher = f
			}
		case elbLogSSLProtocol:
			if f != "-" {
				l.SSLProtocol = f
			}
		}
	}
	return nil
}
