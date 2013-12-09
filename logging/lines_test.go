package logging

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	UNICORN_LINE = `2013-11-09T12:00:00.029051+00:00 eae43d53eaf7 unicorn[55]: [96baf015-c074-4d92-9b5a-d7b9a20e55ca] Started GET "/_status" for 127.0.0.1 at 2013-11-09 12:00:00 +0000`

	SSL_LINE = `2013-11-09T12:00:00.016381+00:00 cnc-618c0f60 ssl_endpoint: 54.248.220.45 - host=195.38.136.104 method=GET status=301 length=184 - total=0.211 upstream_time=- ua="Amazon Route 53 Health Check Service" uri="/_status" ref="some.referer"`

	HAPROXY_LINE = `2013-11-09T12:00:00+00:00 192.168.0.6 haproxy[23201]: 192.168.0.37:54273 [09/Nov/2013:11:59:59.676] in-ff ff/cnc-a6ce4d77:ecb880d6f772:c4093e8b1754 1/2/3/392/395 200 74674 - - ---- 2/4/6/7/8 9/7 "GET /api/v1/categories/3989/photos?param=true&page=1&limit=24 HTTP/1.0"`

	LINE_WITH_SEVERITY = `2013-12-09T09:59:49.290815+00:00 lisa-he-179584 metrix.notice[2835]: net.ip.IncomingPacketsDelivered 1386583189 797103131 host=lisa-he-179584`
)

func TestParseTag(t *testing.T) {
	Convey("parseTags", t, func() {
		So(1, ShouldEqual, 1)
		tags := map[string]SyslogLine{
			"metrix":      {Tag: "metrix"},
			"metrix.info": {Tag: "metrix", Severity: "info"},
			"metrix.info[1234]": {Tag: "metrix", Severity: "info", Pid: 1234},
			"metrix.info[]:": {Tag: "metrix", Severity: "info", Pid: 0},
		}
		for raw, line := range tags {
			tag, severity, pid := parseTag(raw)
			So(tag, ShouldEqual, line.Tag)
			So(severity, ShouldEqual, line.Severity)
			So(pid, ShouldEqual, line.Pid)
		}
	})
}

func TestParseSyslogLine(t *testing.T) {
	Convey("Parse syslog line", t, func() {
		line := &SyslogLine{}
		So(line.Parse(UNICORN_LINE), ShouldBeNil)
		So(line.Host, ShouldEqual, "eae43d53eaf7")
		So(line.Tag, ShouldEqual, "unicorn")
		So(line.Pid, ShouldEqual, 55)
		So(line.Time.Unix(), ShouldEqual, 1383998400)
		So(line.Raw, ShouldEqual, UNICORN_LINE)
	})
}

func TestParseLineWithSeverity(t *testing.T) {
	Convey("Parse line with severity", t, func() {
		line := &SyslogLine{}
		So(line.Parse(LINE_WITH_SEVERITY), ShouldBeNil)
		So(line.Tag, ShouldEqual, "metrix")
		So(line.Severity, ShouldEqual, "notice")
	})
}

func TestUnicornLine(t *testing.T) {
	line := &UnicornLine{}
	Convey("Parse unicorn line", t, func() {
		So(line.Parse(UNICORN_LINE), ShouldBeNil)
		So(line.UUID, ShouldEqual, "96baf015-c074-4d92-9b5a-d7b9a20e55ca")
	})
}

func TestParseSslLine(t *testing.T) {
	line := &NginxLine{}
	Convey("Parse nginx line", t, func() {
		So(line.Parse(SSL_LINE), ShouldBeNil)
		So(line.Host, ShouldEqual, "cnc-618c0f60")
		So(line.Tag, ShouldEqual, "ssl_endpoint")
		So(line.Method, ShouldEqual, "GET")
		So(line.Status, ShouldEqual, "301")
		So(line.Length, ShouldEqual, 184)
		So(line.TotalTime, ShouldEqual, 0.211)
		So(line.UnicornTime, ShouldEqual, 0.0)
		So(line.HttpHost, ShouldEqual, "195.38.136.104")
		So(line.UserAgentName, ShouldEqual, "Amazon Route 53 Health Check Service")
		So(line.Uri, ShouldEqual, "/_status")
		So(line.Referer, ShouldEqual, "some.referer")
	})
}

func TestParseHAProxyLine(t *testing.T) {
	line := &HAProxyLine{}
	Convey("Parse HaProxy line", t, func() {
		So(line.Parse(HAPROXY_LINE), ShouldBeNil)
		So(line.Frontend, ShouldEqual, "in-ff")
		So(line.BackendHost, ShouldEqual, "cnc-a6ce4d77")
		So(line.BackendImageId, ShouldEqual, "ecb880d6f772")
		So(line.BackendContainerId, ShouldEqual, "c4093e8b1754")
		So(line.Status, ShouldEqual, "200")
		So(line.Length, ShouldEqual, 74674)
		So(line.ClientRequestTime, ShouldEqual, 1)
		So(line.ConnectionQueueTime, ShouldEqual, 2)
		So(line.TcpConnectTime, ShouldEqual, 3)
		So(line.ServerResponseTime, ShouldEqual, 392)
		So(line.SessionDurationTime, ShouldEqual, 395)
		So(line.ActiveConnections, ShouldEqual, 2)
		So(line.FrontendConnections, ShouldEqual, 4)
		So(line.BackendConnectons, ShouldEqual, 6)
		So(line.ServerConnections, ShouldEqual, 7)
		So(line.Retries, ShouldEqual, 8)
		So(line.ServerQueue, ShouldEqual, 9)
		So(line.BackendQueue, ShouldEqual, 7)
		So(line.Method, ShouldEqual, "GET")
		So(line.Uri, ShouldEqual, "/api/v1/categories/3989/photos?param=true&page=1&limit=24")
	})
}
