package dockerclient

import "encoding/json"

type Version struct {
	Arch          string
	GitCommit     string
	GoVersion     string
	KernelVersion string
	Os            string
	Version       string
}

func (dh *DockerHost) ServerVersion() (*Version, error) {
	if dh.cachedServerVersion != nil {
		return dh.cachedServerVersion, nil
	}

	rsp, e := dh.httpClient.Get(dh.url() + "/version")
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
