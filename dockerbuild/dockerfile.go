package dockerbuild

import (
	"regexp"
)

type Dockerfile []byte

func NewDockerfile(b []byte) Dockerfile {
	return b
}

var proxyRegexp = regexp.MustCompile("^(FROM .*?)\n")

func (df Dockerfile) MixinProxy(proxy string) Dockerfile {
	s := string(df)
	m := proxyRegexp.FindStringSubmatch(s)
	if len(m) > 1 {
		s = proxyRegexp.ReplaceAllString(s, m[1]+"\nENV http_proxy "+proxy + "\n")
	}
	return NewDockerfile([]byte(s))
}
