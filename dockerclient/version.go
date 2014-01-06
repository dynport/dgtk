package dockerclient

import (
	"encoding/json"
	"github.com/dynport/dgtk/version"
	"net/http"
)

type Version struct {
	Arch          string
	GitCommit     string
	GoVersion     string
	KernelVersion string
	Os            string
	Version       string
}

var jsonBuildStreamFrom *version.Version

func init() {
	var e error
	jsonBuildStreamFrom, e = version.Parse("0.7.0")
	if e != nil {
		panic(e.Error())
	}
}

func (v *Version) jsonBuildStream() bool {
	parsed, e := version.Parse(v.Version)
	if e != nil {
		panic(e.Error())
	}
	return !parsed.Less(jsonBuildStreamFrom)
}
func mustBe(b bool, e error) bool {
	if e != nil {
		panic(e.Error())
	}
	return b
}

func (dh *DockerHost) ServerVersion() (*Version, error) {
	if dh.cachedServerVersion != nil {
		return dh.cachedServerVersion, nil
	}
	rsp, e := http.Get(dh.url() + "/version")
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	v := &Version{}
	e = json.NewDecoder(rsp.Body).Decode(v)
	if e != nil {
		dh.cachedServerVersion = v
	}
	return v, e
}
