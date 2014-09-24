package dpr

type ContainerConfig struct {
	Tty          bool
	Cmd          interface{}
	Env          interface{}
	Image        string
	Hostname     string
	StdinOnce    bool
	AttachStdin  bool
	User         string
	PortSpecs    interface{}
	Memory       int64
	MemorySwap   int64
	AttachStderr bool
	AttachStdout bool
	OpenStdin    bool
}
