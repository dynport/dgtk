package logging

import "testing"

func TestAWSELBLogHTTP(t *testing.T) {
	line := `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000073 0.001048 0.000057 200 200 0 29 "GET http://www.example.com:80/ HTTP/1.1" "curl/7.38.0" - -`

	l := &ElasticLoadBalancerLog{}
	if err := l.Load(line); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"ELB", "my-loadbalancer", l.Elb},
		{"Method", "GET", l.Method},
		{"URL", "http://www.example.com:80/", l.Url},
		{"UserAgent", "curl/7.38.0", l.UserAgent},
		{"SSLCipher", "", l.SSLCipher},
		{"SSLProtocol", "", l.SSLProtocol},
		{"BackendProcessingTime", 0.001048, l.BackendProcessingTime},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}

func TestAWSELBLogHTTPS(t *testing.T) {
	line := `2015-05-13T23:39:43.945958Z my-loadbalancer 192.168.131.39:2817 10.0.0.1:80 0.000086 0.001048 0.001337 200 200 0 57 "GET https://www.example.com:443/ HTTP/1.1" "curl/7.38.0" DHE-RSA-AES128-SHA TLSv1.2`
	l := &ElasticLoadBalancerLog{}
	if err := l.Load(line); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"ELB", "my-loadbalancer", l.Elb},
		{"Method", "GET", l.Method},
		{"URL", "https://www.example.com:443/", l.Url},
		{"UserAgent", "curl/7.38.0", l.UserAgent},
		{"BackendProcessingTime", 0.001048, l.BackendProcessingTime},
		{"SSLCipher", "DHE-RSA-AES128-SHA", l.SSLCipher},
		{"SSLProtocol", "TLSv1.2", l.SSLProtocol},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}
