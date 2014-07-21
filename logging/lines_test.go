package logging

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	UNICORN_LINE = `2013-11-09T12:00:00.029051+00:00 eae43d53eaf7 unicorn[55]: [96baf015-c074-4d92-9b5a-d7b9a20e55ca] Started GET "/_status" for 127.0.0.1 at 2013-11-09 12:00:00 +0000`

	SSL_LINE = `2013-11-09T12:00:00.016381+00:00 cnc-618c0f60 ssl_endpoint: 54.248.220.45 - host=195.38.136.104 method=GET status=301 length=184 - total=0.211 upstream_time=- ua="Amazon Route 53 Health Check Service" uri="/_status" ref="some.referer"`

	HAPROXY_LINE = `2013-11-09T12:00:00+00:00 192.168.0.6 haproxy[23201]: 192.168.0.37:54273 [09/Nov/2013:11:59:59.676] in-ff ff/cnc-a6ce4d77:ecb880d6f772:c4093e8b1754 1/2/3/392/395 200 74674 - - ---- 2/4/6/7/8 9/7 "GET /api/v1/categories/3989/photos?param=true&page=1&limit=24 HTTP/1.0"`

	LINE_WITH_SEVERITY        = `2013-12-09T09:59:49.290815+00:00 lisa-he-179584 metrix.notice[2835]: net.ip.IncomingPacketsDelivered 1386583189 797103131 host=lisa-he-179584`
	LINE_WITH_KEY_VALUE_PAIRS = `2013-12-09T14:19:14.575268+01:00 some.host nginx.notice[]: some.ip - host=phraseapp.com method=GET status=200 length=11928 pid=24969 rev=db7b58fa06cc uuid=56be52ae-7cd7-4a9e-b7ce-4a55074976ad action=translations#index etag=179d6a6dccacc3116cf3beb49187ff3d rack=2.120354 redis=0.000666/1 db=1.001724/240 solr=0.147277/23 db_cache=0.131241/473 total=2.235 upstream_time=2.235 ua="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36" uri="/projects/phraseapp-demo-some-id/locales/de/translations" ref="https://phraseapp.com/en/account/login"`
)

func TestFields(t *testing.T) {
	Convey("Fields", t, func() {
		line := `this is a test`

		fields := Fields(line)
		So(len(fields), ShouldEqual, 4)
		So(fields[1], ShouldEqual, "is")

		line = `this "is a test" with extras`
		fields = Fields(line)
		So(len(fields), ShouldEqual, 4)
		So(fields[1], ShouldEqual, "is a test")
		So(fields[2], ShouldEqual, "with")

		line = `this is us="user agent" test`
		fields = Fields(line)
		So(len(fields), ShouldEqual, 4)
		So(fields[0], ShouldEqual, "this")
		So(fields[2], ShouldEqual, `us=user agent`)
	})
}

func TestLineTags(t *testing.T) {
	line := &SyslogLine{}
	Convey("Line Tags", t, func() {
		So(line.Parse(LINE_WITH_KEY_VALUE_PAIRS), ShouldBeNil)
		tags := line.Tags()
		So(len(tags), ShouldBeGreaterThan, 1)
		So(tags["length"], ShouldEqual, 11928)
		So(tags["solr_time"], ShouldEqual, 0.147277)
		So(tags["solr_calls"], ShouldEqual, 23)

		So(tags["redis_time"], ShouldEqual, 0.000666)
		So(tags["redis_calls"], ShouldEqual, 1)
		So(tags["ua"], ShouldEqual, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36")
		So(tags["uri"], ShouldEqual, `/projects/phraseapp-demo-some-id/locales/de/translations`)

		tags = parseTags(HAPROXY_LINE)
		So(len(tags), ShouldEqual, 0)
	})

	Convey("line with multiple forwardings", t, func() {
		raw := `2013-11-20T16:00:00.364664+00:00 1fb6092433dc nginx: 192.168.0.6 176.199.77.195, 10.22.61.117, 2.22.61.87 host=www.1414.de`
		line := &NginxLine{}
		So(line.Parse(raw), ShouldBeNil)
		So(line.Host, ShouldEqual, "1fb6092433dc")
		So(strings.Join(line.XForwardedFor, " "), ShouldEqual, "192.168.0.6 176.199.77.195 10.22.61.117 2.22.61.87")
		So(line.Message, ShouldEqual, "192.168.0.6 176.199.77.195, 10.22.61.117, 2.22.61.87 host=www.1414.de")

	})

	Convey("parseTagValue", t, func() {
		m := map[string]interface{}{
			"200":   200,
			"2.235": 2.235,
			"test":  "test",
		}
		for from, to := range m {
			So(parseTagValue(from), ShouldEqual, to)
		}
	})
}

func TestParseTag(t *testing.T) {
	Convey("parseTags", t, func() {
		tags := map[string]SyslogLine{
			"metrix":               {Tag: "metrix"},
			"metrix.info":          {Tag: "metrix", Severity: "info"},
			"metrix.info[1234]":    {Tag: "metrix", Severity: "info", Pid: 1234},
			"metrix.info[]:":       {Tag: "metrix", Severity: "info", Pid: 0},
			"mongod.27017.warning": {Port: "27017", Tag: "mongod", Severity: "warning", Pid: 0},
		}
		for raw, line := range tags {
			tag, port, severity, pid := parseTag(raw)
			So(tag, ShouldEqual, line.Tag)
			So(severity, ShouldEqual, line.Severity)
			So(pid, ShouldEqual, line.Pid)
			So(port, ShouldEqual, line.Port)
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
		So(line.Port, ShouldEqual, "")
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
