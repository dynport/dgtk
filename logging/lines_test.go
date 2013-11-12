package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const UNICORN_LINE = `2013-11-09T12:00:00.029051+00:00 eae43d53eaf7 unicorn[55]: [96baf015-c074-4d92-9b5a-d7b9a20e55ca] Started GET "/_status" for 127.0.0.1 at 2013-11-09 12:00:00 +0000`

const SSL_LINE = `2013-11-09T12:00:00.016381+00:00 cnc-618c0f60 ssl_endpoint: 54.248.220.45 - host=195.38.136.104 method=GET status=301 length=184 - total=0.211 upstream_time=- ua="Amazon Route 53 Health Check Service" uri="/_status" ref="some.referer"`

const HAPROXY_LINE = `2013-11-09T12:00:00+00:00 192.168.0.6 haproxy[23201]: 192.168.0.37:54273 [09/Nov/2013:11:59:59.676] in-ff ff/cnc-a6ce4d77:ecb880d6f772:c4093e8b1754 1/2/3/392/395 200 74674 - - ---- 2/4/6/7/8 9/7 "GET /api/v1/categories/3989/photos?param=true&page=1&limit=24 HTTP/1.0"`

func TestParseSyslogLine(t *testing.T) {
	line := &SyslogLine{}
	assert.Nil(t, line.Parse(UNICORN_LINE))
	assert.Equal(t, line.Host, "eae43d53eaf7", "Host")
	assert.Equal(t, line.Tag, "unicorn", "Tag")
	assert.Equal(t, line.Pid, 55, "Pid")
	assert.Equal(t, line.Time.Unix(), 1383998400)
	assert.Equal(t, line.Raw, UNICORN_LINE)
}

func TestUnicornLine(t *testing.T) {
	line := &UnicornLine{}
	assert.Nil(t, line.Parse(UNICORN_LINE))
	assert.Equal(t, line.UUID, "96baf015-c074-4d92-9b5a-d7b9a20e55ca")
}

func TestParseSslLine(t *testing.T) {
	line := &NginxLine{}
	assert.Nil(t, line.Parse(SSL_LINE))
	assert.Equal(t, line.Host, "cnc-618c0f60")
	assert.Equal(t, line.Tag, "ssl_endpoint")
	assert.Equal(t, line.Method, "GET")
	assert.Equal(t, line.Status, "301")
	assert.Equal(t, line.Length, 184)
	assert.Equal(t, line.TotalTime, 0.211)
	assert.Equal(t, line.UnicornTime, 0.0)
	assert.Equal(t, line.HttpHost, "195.38.136.104")
	assert.Equal(t, line.UserAgentName, "Amazon Route 53 Health Check Service")
	assert.Equal(t, line.Uri, "/_status")
	assert.Equal(t, line.Referer, "some.referer")
}

func TestParseHAProxyLine(t *testing.T) {
	line := &HAProxyLine{}
	assert.Nil(t, line.Parse(HAPROXY_LINE))
	assert.Equal(t, line.Frontend, "in-ff")
	assert.Equal(t, line.BackendHost, "cnc-a6ce4d77")
	assert.Equal(t, line.BackendImageId, "ecb880d6f772")
	assert.Equal(t, line.BackendContainerId, "c4093e8b1754")
	assert.Equal(t, line.Status, "200")
	assert.Equal(t, line.Length, 74674)
	assert.Equal(t, line.ClientRequestTime, 1)
	assert.Equal(t, line.ConnectionQueueTime, 2)
	assert.Equal(t, line.TcpConnectTime, 3)
	assert.Equal(t, line.ServerResponseTime, 392)
	assert.Equal(t, line.SessionDurationTime, 395)
	assert.Equal(t, line.ActiveConnections, 2)
	assert.Equal(t, line.FrontendConnections, 4)
	assert.Equal(t, line.BackendConnectons, 6)
	assert.Equal(t, line.ServerConnections, 7)
	assert.Equal(t, line.Retries, 8)
	assert.Equal(t, line.ServerQueue, 9)
	assert.Equal(t, line.BackendQueue, 7)
	assert.Equal(t, line.Method, "GET")
	assert.Equal(t, line.Uri, "/api/v1/categories/3989/photos?param=true&page=1&limit=24")

}
