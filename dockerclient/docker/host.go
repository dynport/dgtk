package docker

type Port string

type HostConfig struct {
	Binds        []string
	PortBindings map[Port][]PortBinding
	LxcConf      map[string]string
}

type PortBinding struct {
	HostIp   string
	HostPort string
}
