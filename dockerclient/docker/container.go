package docker

import (
	"fmt"
	"strings"
	"time"
)

type NetworkSettings struct {
	Ip          string `json:"IPAddress"`
	IPPrefixLen int
	Gateway     string
	Bridge      string
	PortMapping map[string]map[string]string
	Ports       map[Port][]PortBinding
}

func (settings *NetworkSettings) PortMappingString() string {
	out := []string{}
	tcp := []string{}
	udp := []string{}
	if mappings := settings.TcpPortMapping(); len(mappings) > 0 {
		for v, k := range mappings {
			tcp = append(tcp, k+"=>"+v)
		}
		out = append(out, strings.Join(tcp, ", "))
	}
	if mappings := settings.UdpPortMapping(); len(mappings) > 0 {
		for v, k := range mappings {
			udp = append(udp, k+"=>"+v)
		}
		out = append(out, "UDP: "+strings.Join(udp, ", "))
	}
	return strings.Join(out, " | ")
}

func (settings *NetworkSettings) TcpPortMapping() map[string]string {
	return settings.PortMappingFor("Tcp")
}

func (settings *NetworkSettings) UdpPortMapping() map[string]string {
	return settings.PortMappingFor("Udp")
}

func (settings *NetworkSettings) PortMappingFor(protocol string) map[string]string {
	if mapping := settings.PortMapping[protocol]; mapping != nil {
		return mapping
	}
	return map[string]string{}
}

type PortConfig struct {
	Private  int    `json:"PrivatePort"`
	Public   int    `json:"PublicPort"`
	Protocol string `json:"Type"`
	Ip       string `json:"IP"`
}

type Container struct {
	Id         string
	Image      string
	Command    string
	Created    int64 `json:"Created"`
	Status     string
	Ports      []*PortConfig
	SizeRw     int
	SizeRootFs int
	Names      []string
}

func (self *Container) CreatedAt() time.Time {
	return time.Unix(self.Created, 0)
}

func (container *Container) String() string {
	return fmt.Sprintf("%s: %s", container.Id, container.Status)
}

type ContainerInfo struct {
	Id              string `json:"ID"`
	Image           string
	created         time.Time `json:"Created"`
	SysInitPath     string
	ResolvConfPath  string
	Volumes         map[string]string
	VolumesRW       map[string]string
	Path            string
	Args            []string
	ContainerConfig ContainerConfig `json:"Config"`
	NetworkConfig   NetworkSettings `json:"NetworkSettings"`
	HostConfig      HostConfig      `json:"HostConfig"`
}

func (self *ContainerInfo) CreatedAt() time.Time {
	return self.created
}

// https://github.com/dotcloud/docker/blob/master/container.go#L60-81
type ContainerConfig struct {
	Hostname        string
	Domainname      string
	User            string
	Memory          int64 // Memory limit (in bytes)
	MemorySwap      int64 // Total memory usage (memory + swap); set `-1' to disable swap
	CpuShares       int64 // CPU shares (relative weight vs. other containers)
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	PortSpecs       []string
	ExposedPorts    map[Port]struct{}
	Tty             bool // Attach standard streams to a tty, including stdin if it is not closed.
	OpenStdin       bool // Open stdin
	StdinOnce       bool // If true, close stdin after the 1 attached client disconnects.
	Env             []string
	Cmd             []string
	Dns             []string
	Image           string // Name of the image as it was passed by the operator (eg. could be symbolic)
	Volumes         map[string]struct{}
	VolumesFrom     string
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
}
