package wunderproxy

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dynport/dgtk/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/dockerclient"
	"github.com/dynport/dgtk/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/dockerclient/docker"
	"github.com/dynport/dgtk/wunderproxy/Godeps/_workspace/src/github.com/dynport/gocloud/aws/s3"
)

type ContainerManager struct {
	S3Bucket     string
	S3Prefix     string
	Proxy        *Proxy
	RegistryPort int

	AppName string

	dockerHost *dockerclient.DockerHost
	s3client   *s3.Client

	containerEnv     map[string]string
	containers       map[string]int
	currentContainer string
}

func NewContainerManager(s3Bucket, s3Prefix, appname string, proxy *Proxy, cfgFile string, rport int) (*ContainerManager, error) {
	cm := &ContainerManager{
		S3Bucket:     s3Bucket,
		S3Prefix:     s3Prefix,
		AppName:      appname,
		Proxy:        proxy,
		RegistryPort: rport,
		s3client:     s3.NewFromEnv(),
		containers:   map[string]int{},
		containerEnv: map[string]string{},
	}

	if cfgFile != "" {
		err := cm.readConfigFile(cfgFile)
		if err != nil {
			return nil, err
		}
	}

	var e error
	cm.dockerHost, e = dockerclient.New("127.0.0.1", 4243)
	if e != nil {
		return nil, e
	}

	return cm, func() error {
		err := cm.runLatest()
		switch err {
		case nil, ErrorLaunchConfigNotFound:
			return nil
		default:
			return err
		}
	}()
}

func (cm *ContainerManager) readConfigFile(path string) error {
	fh, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open container config file: %s", err)
	}
	defer fh.Close()

	sc := bufio.NewScanner(fh)
	for sc.Scan() {
		line := sc.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("syntax error: expected lines of key=value format, got\n\t%q", line)
		}
		k, v := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		cm.containerEnv[k] = v
	}
	return sc.Err()
}

func (cm *ContainerManager) runLatest() error {
	if cm.s3client == nil {
		return fmt.Errorf("S3 client not configured")
	}

	cfg, err := LoadCurrentLaunchConfig(cm.s3client, cm.S3Bucket, cm.S3Prefix)
	switch {
	case err != nil:
		return err
	case cfg == nil: // no current launch configuration
		return nil
	}

	var port int
	cm.currentContainer, port, err = cm.execute(cfg)
	if err != nil {
		return err
	}

	cm.containers[cfg.Revision] = port
	cm.Proxy.Update(fmt.Sprintf("localhost:%d", port), cfg.MaintenancePath)
	return cm.StopOldContainers()
}

func (cm *ContainerManager) StartContainer(hash string) (string, int, error) {
	cm.s3client = s3.NewFromEnv()

	cfg, e := LoadLaunchConfig(cm.s3client, cm.S3Bucket, cm.S3Prefix, hash)
	if e != nil {
		return "", -1, e
	}

	containerId, port, err := cm.execute(cfg)
	if err != nil {
		return "", -1, err
	}

	cm.currentContainer = containerId
	cm.containers[hash] = port

	return containerId, port, nil
}

func (cm *ContainerManager) SwitchContainer(hash string) (string, int, error) {
	cm.s3client = s3.NewFromEnv()

	port, ok := cm.containers[hash]
	if !ok {
		return "", -1, fmt.Errorf("don't know about a container with revision %q", hash)
	}

	cfg, e := LoadLaunchConfig(cm.s3client, cm.S3Bucket, cm.S3Prefix, hash)
	if e != nil {
		return "", -1, e
	}

	cm.Proxy.Update(fmt.Sprintf("localhost:%d", port), cfg.MaintenancePath)

	// Stop old containers. Ignore errors ... they'll get visible soon enough.
	go func() {
		logger.Printf("stopping old containers")
		err := cm.StopOldContainers()
		if err != nil {
			logger.Printf("error stopping old containers: %s", err)
		}
	}()

	return cm.currentContainer, port, nil
}

func (cm *ContainerManager) execute(cfg *LaunchConfig) (string, int, error) {
	var containerId string
	var port int
	e := func() error {
		container, err := cm.containerForImage(cfg.ContainerConfig.Image)
		if err != nil {
			return err
		}

		if container != nil {
			containerId = container.Id
			switch {
			case strings.HasPrefix(container.Status, "Up "):
				logger.Printf("container with id %s runs the required image", containerId)
				port, err = cm.waitForContainer(containerId, cfg.HealthCheckPath)
			default:
				logger.Printf("starting existing container %q", containerId)
				containerId, port, err = cm.startAndWaitForContainer(containerId, cfg)
			}
		} else {
			logger.Printf("creating new container")
			containerId, port, err = cm.createAndWaitForContainer(cfg)
		}
		return err
	}()

	return containerId, port, e
}

func (cm *ContainerManager) containerForImage(imageId string) (*docker.Container, error) {
	containers, err := cm.availableContainers()
	if err != nil {
		return nil, err
	}

	for i := range containers {
		if containers[i].Image == imageId {
			return containers[i], nil
		}
	}

	return nil, nil
}

func (cm *ContainerManager) availableContainers() (containers []*docker.Container, e error) {
	imagePrefix := fmt.Sprintf("localhost:%d/%s:", cm.RegistryPort, cm.AppName)

	availCont, e := cm.dockerHost.ListContainers(&dockerclient.ListContainersOptions{All: true})
	if e != nil {
		return nil, e
	}

	for i := range availCont {
		imageId := availCont[i].Image

		if strings.HasPrefix(imageId, imagePrefix) {
			revision := "/" + strings.TrimPrefix(imageId, imagePrefix)
			containerNames := availCont[i].Names
			for j := range containerNames {
				if containerNames[j] == revision {
					containers = append(containers, availCont[i])
					break
				}
			}
		}
	}
	return containers, nil
}

func (cm *ContainerManager) createAndWaitForContainer(cfg *LaunchConfig) (containerId string, port int, e error) {
	for k, v := range cm.containerEnv {
		cfg.ContainerConfig.Env = append(cfg.ContainerConfig.Env, fmt.Sprintf("%s=%v", k, v))
	}

	containerId, e = cm.dockerHost.CreateContainer(cfg.ContainerConfig, cfg.Revision)
	if e != nil {
		return "", 0, e
	}
	return cm.startAndWaitForContainer(containerId, cfg)
}

func (cm *ContainerManager) startAndWaitForContainer(containerId string, cfg *LaunchConfig) (_ string, port int, e error) {
	e = cm.dockerHost.StartContainer(containerId, cfg.HostConfig)
	if e != nil {
		return "", 0, e
	}
	logger.Printf("started container with image %q", cfg.ContainerConfig.Image)
	port, e = cm.waitForContainer(containerId, cfg.HealthCheckPath)
	return containerId, port, e
}

func (cm *ContainerManager) waitForContainer(containerId, healthCheckPath string) (port int, e error) {
	port, e = cm.exposedContainerPort(containerId, "9292/tcp")
	if e != nil {
		return 0, e
	}

	for running := true; running; {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/%s", port, healthCheckPath), nil)
		if err != nil {
			return 0, err
		}
		req.Header.Add("X-Forwarded-Proto", "https") // be sure to succeed even for hosts with SSL required
		resp, e := http.DefaultClient.Do(req)
		if e == nil {
			resp.Body.Close()
		}
		if e == nil && resp.StatusCode == 200 {
			break
		}
		logger.Printf("waiting another 5 seconds")
		time.Sleep(5 * time.Second)
	}

	return port, nil
}

func (cm *ContainerManager) exposedContainerPort(containerId string, srcPort docker.Port) (int, error) {
	cinfo, e := cm.dockerHost.Container(containerId)
	if e != nil {
		return 0, e
	}

	destPorts := cinfo.NetworkConfig.Ports[srcPort]
	if len(destPorts) != 1 {
		return 0, fmt.Errorf("panic not knowing what to do!") // TODO clean up
	}

	return strconv.Atoi(destPorts[0].HostPort)
}

func (cm *ContainerManager) StopOldContainers() error {
	containers, err := cm.availableContainers()
	if err != nil {
		return err
	}

	for i := range containers {
		if containers[i].Id == cm.currentContainer {
			continue
		}

		logger.Printf("stopping old container %s", containers[i].Id)
		err := cm.dockerHost.StopContainer(containers[i].Id)
		if err != nil {
			return err
		}

		err = cm.dockerHost.RemoveContainer(containers[i].Id)
		if err != nil {
			return err
		}

		err = cm.dockerHost.DeleteImage(containers[i].Image)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cm *ContainerManager) Stats(stats map[string]interface{}) error {
	cStat, err := cm.dockerHost.HostInfo()
	if err != nil {
		return err
	}

	stats["DockerState"] = "Ok"
	if cStat.Driver != "aufs" {
		stats["DockerState"] = fmt.Sprintf("Wrong driver; expected %q got %q", "aufs", cStat.Driver)
	}

	return nil
}
