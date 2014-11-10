package logging

import (
	"strings"
	"testing"
)

func TestParseNginx(t *testing.T) {
	r := `2014-11-07T14:16:59.014363+00:00 i-fa2e28b9 nginx.notice[]: 172.31.6.126 164.177.8.109 host=phraseapp.com method=GET status=200 length=33 pid=15076 rev=7bfcf0a659ae uuid=3742facc-a8cd-4cd0-aa63-b4e69b58e20e action=translations#placeholders etag=d751713988987e9331980363e24189ce rack=0.016725 db=0.005059/7 total=0.019 upstream_time=0.019 ua="Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.111 Safari/537.36" uri="/some/path" ref="/some/ref"`

	l := &NginxLine{}
	err := l.Parse(r)
	if err != nil {
		t.Fatalf("erorr parsing nginx line")
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"IPs", "172.31.6.126,164.177.8.109", strings.Join(l.XForwardedFor, ",")},
		{"Host", "i-fa2e28b9", l.Host},
		{"UserAgent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.111 Safari/537.36", l.UserAgentName},
		{"upstream_time", "200", l.Status},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}

}
