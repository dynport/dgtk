package docker

type Port string

type HostConfig struct {
	PublishAllPorts bool
	NetworkMode     string
	Binds           []string
	PortBindings    map[Port][]PortBinding
	LxcConf         interface{}
}

type PortBinding struct {
	HostIp   string
	HostPort string
}
