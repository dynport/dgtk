package wunderproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/dynport/dgtk/dockerclient/docker"
)

type Client struct {
	Addr  string
	Debug bool
}

func (c *Client) Container(id string) (cont *Container, err error) {
	return cont, c.loadJSON("/containers/"+id+"/json", &cont)
}

func (c *Client) HostInfo() (info *HostInfo, err error) {
	return info, c.loadJSON("/info", &info)
}

func (c *Client) ListContainers(all bool) (list []*Container, err error) {
	return list, c.loadJSON("/containers/json?all="+fmt.Sprintf("%t", all), &list)
}

type Image struct {
	Id string
}

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
	VolumesFrom     interface{} `json:"VolumesFrom,omitempty"`
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	HostConfig      *HostConfig
}

func (c *Client) exposedContainerPort(containerId string, srcPort docker.Port) (int, error) {
	cinfo, e := c.Container(containerId)
	if e != nil {
		return 0, e
	}

	log.Printf("%#v", cinfo.NetworkConfig.Ports)
	destPorts := cinfo.NetworkConfig.Ports[srcPort]
	if len(destPorts) == 1 {
		return strconv.Atoi(destPorts[0].HostPort)
	}
	destPorts = cinfo.NetworkSettings.Ports[srcPort]
	if len(destPorts) != 1 {
		return 0, fmt.Errorf("panic not knowing what to do!") // TODO clean up
	}
	return strconv.Atoi(destPorts[0].HostPort)
}

func (c *Client) CreateContainer(cfg *ContainerConfig, name string) (string, error) {
	// set image
	var img *Image
	if img != nil {
		cfg.Env = append(cfg.Env, "DOCKER_IMAGE="+img.Id)
	}
	var cnt *Container
	p := "containers/create"
	if name != "" {
		p += "?" + (url.Values{"name": {name}}).Encode()
	}
	err := c.postJSON(p, cfg, &cnt)
	if err != nil {
		return "", err
	}
	return cnt.Id, nil
}

func (c *Client) postJSON(p string, i interface{}, res interface{}) error {
	u := "http://127.0.0.1:4243" + path.Join("/", p)
	log.Printf("url=%q", u)
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	rsp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return json.NewDecoder(rsp.Body).Decode(&res)
}

func (c *Client) StartContainer(id string) error {
	rsp, err := c.client().Post("http://127.0.0.1:4243/containers/"+id+"/start", "application/json", nil)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return nil
}

func (c *Client) StopContainer(id string) error {
	return dockerExec("stop", id)
}

// wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/dockerclient/container.go

func (c *Client) RunContiner(id, hostConfig interface{}) error {
	panic("implement me")
}

func (c *Client) RemoveContainer(id string) error {
	return dockerExec("rm", id)
}

func (c *Client) DeleteImage(id string) error {
	return dockerExec("rmi", id)
}

func dockerExec(args ...string) error {
	b, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", err, string(b))
	}
	return nil
}

func (c *Client) client() *http.Client {
	if c.Addr != "" {
		return http.DefaultClient

	}
	tr := &http.Transport{}
	tr.Dial = func(nw, addr string) (net.Conn, error) {
		return net.Dial("unix", "/var/run/docker.sock")
	}
	cl := http.DefaultClient
	cl.Transport = tr
	return cl
}

func (c *Client) loadJSON(p string, i interface{}) error {
	u := "http://127.0.0.1:4243" + path.Join("/", p)
	rsp, err := c.client().Get(u)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	var r io.Reader = rsp.Body
	if c.Debug {
		r = io.TeeReader(r, os.Stdout)
	}
	return json.NewDecoder(r).Decode(&i)
}
