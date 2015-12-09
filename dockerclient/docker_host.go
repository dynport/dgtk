package dockerclient

type DockerHostInfo struct {
	Containers         int
	Images             int
	Driver             string
	DriverStatus       [][]string
	ExecutionDriver    string
	KernelVersion      string
	NCPU               int
	MemTotal           int
	Name               string
	ID                 string
	NFd                int
	NGoroutines        int
	NEventsListener    int
	InitPath           string
	InitSha1           string
	IndexServerAddress string
	MemoryLimit        int
	SwapLimit          int
	Labels             []string
	DockerRootDir      string
	OperatingSystem    string
}

func (dh *Client) HostInfo() (*DockerHostInfo, error) {
	u := dh.Address + "/info"

	dhi := &DockerHostInfo{}
	return dhi, dh.getJSON(u, &dhi)
}
