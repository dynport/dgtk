package docker

type HostConfig struct {
	Binds   []string
	LxcConf map[string]string
}
