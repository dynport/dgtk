package dockerclient

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"github.com/dynport/dgtk/dockerclient/docker"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var (
	imageIdRegexp   = regexp.MustCompile("Successfully built (.*)")
	imageNameRegexp = regexp.MustCompile("([\\w.]+(?::\\d+)?/)?(\\w+)(:[\\w.]+)?")
)

// Get the list of all images available on the this host.
func (dh *DockerHost) Images() (images []*docker.Image, e error) {
	e = dh.getJSON(dh.url()+"/images/json", &images)
	return images, e
}

// Get the details for image with the given id (either hash or name).
func (dh *DockerHost) ImageDetails(id string) (imageDetails *docker.ImageDetails, e error) {
	imageDetails = &docker.ImageDetails{}
	e = dh.getJSON(dh.url()+"/images/"+id+"/json", imageDetails)
	return imageDetails, e
}

// Get the given image's history.
func (dh *DockerHost) ImageHistory(id string) (imageHistory *docker.ImageHistory, e error) {
	imageHistory = &docker.ImageHistory{}
	e = dh.getJSON(dh.url()+"/images/"+id+"/history", imageHistory)
	return imageHistory, e
}

// Create a new image from the given dockerfile. If name is non empty the new image is named accordingly. If a writer is
// given it is used to send the docker output to.
func (dh *DockerHost) BuildImage(dockerfile, tag string) (imageId string, e error) {
	buf, e := dh.createDockerfileArchive(dockerfile)
	if e != nil {
		return
	}

	url := dh.url() + "/build?"
	if tag != "" {
		url += "t=" + tag
	}

	rsp, e := dh.httpClient.Post(url, "application/tar", buf)
	if e != nil {
		return
	}
	defer rsp.Body.Close()

	if !success(rsp) {
		return "", fmt.Errorf("failed to send command to %q: %s", url, rsp.Status)
	}

	scanner := bufio.NewScanner(rsp.Body)
	var last string
	for scanner.Scan() {
		last = scanner.Text()
		dh.Logger.Debug(last)
	}

	s := imageIdRegexp.FindStringSubmatch(string(last))
	if len(s) != 2 {
		return "", fmt.Errorf("unable to extract image id from response: %q", last)
	}
	imageId = s[1]
	return imageId, nil
}

// Tag the image with the given repository and tag. The tag is optional.
func (dh *DockerHost) TagImage(imageId, repository, tag string) (e error) {
	if repository == "" {
		return fmt.Errorf("empty repository given")
	}
	url := dh.url() + "/images/" + imageId + "/tag?repo=" + repository

	if tag != "" {
		url += "&tag=" + tag
	}
	rsp, e := dh.post(url)
	defer rsp.Body.Close()
	return e
}

// Pull the given image from the registry (part of the image name).
func (dh *DockerHost) PullImage(name string) error {
	if name == "" {
		return fmt.Errorf("no image name given")
	}

	registry, repository, tag := splitImageName(name)

	reqUrl := dh.url() + "/images/create"
	values := &url.Values{}
	values.Add("fromImage", registry+"/"+repository)
	values.Add("repo", repository)
	if registry != "" {
		values.Add("registry", registry)
	}
	if tag != "" {
		values.Add("tag", tag)
	}

	rsp, e := dh.post(reqUrl + "?" + values.Encode())
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	if !success(rsp) {
		return fmt.Errorf("failed to fetch image")
	}

	messages, e := splitDockerStatusMessages(rsp.Body)
	if e != nil {
		return e
	}

	for i := range messages {
		if messages[i].Error != "" {
			return fmt.Errorf("failed to pull image: %s", messages[i].Error)
		}
	}

	return dh.waitForTag(registry+"/"+repository, tag, 10)
}

// Push the given image to the registry. The name should be <registry>/<repository>.
func (dh *DockerHost) PushImage(name string) error {
	if name == "" {
		return fmt.Errorf("no image name given")
	}
	registry, _, _ := splitImageName(name)
	if registry == "" {
		return fmt.Errorf("no registry given")
	}

	dh.Logger.Infof("pushing image %s to registry %s", name, registry)
	buf := &bytes.Buffer{}
	buf.WriteString(FAKE_AUTH)
	url := dh.url() + "/images/" + name + "/push?registry=" + registry

	rsp, e := dh.postWithBuffer(url, buf)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()
	if !success(rsp) {
		scanner := bufio.NewScanner(rsp.Body)
		for scanner.Scan() {
			dh.Logger.Debug(scanner.Text())
		}
		return fmt.Errorf("failed to push image:", rsp.Status)
	}
	return nil
}

// Delete the given image from the docker host.
func (dh *DockerHost) DeleteImage(name string) error {
	if name == "" {
		return fmt.Errorf("no image name given")
	}

	req, e := http.NewRequest("DELETE", dh.url()+"/images/"+name, nil)
	if e != nil {
		return e
	}

	resp, e := dh.httpClient.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()

	if !success(resp) {
		return fmt.Errorf("failed to delete image %s", name)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		dh.Logger.Debug(line)
	}

	return nil
}

func (self *DockerHost) createDockerfileArchive(dockerfile string) (buf *bytes.Buffer, e error) {
	body := []byte(dockerfile)
	buf = new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	header := &tar.Header{Name: "Dockerfile", Size: int64(len(body))}
	if e = tw.WriteHeader(header); e != nil {
		return nil, e
	}
	if _, e = tw.Write(body); e != nil {
		return nil, e
	}
	if e = tw.Close(); e != nil {
		return nil, e
	}
	return buf, nil
}

func (dh *DockerHost) waitForTag(repository, tag string, timeout int) error {
	for {
		dh.Logger.Debug("waiting for tag", tag)
		imageDetails, e := dh.ImageDetails(repository + ":" + tag)
		if e != nil {
			if e.Error() == "resource not found" {
				dh.Logger.Debug("got not found, waiting")
				time.Sleep(1 * time.Second)
				continue
			}
			return e
		}
		dh.Logger.Debug("got image details:", imageDetails)
		return nil
	}
}

// Every image is named after the following pattern:
//   <registry>/<repository>:<tag>
// with registry being of the form "<hostname>:<port>" and repository being a string of [A-Za-z0-9_].
func splitImageName(name string) (registry, repository, tag string) {
	s := imageNameRegexp.FindStringSubmatch(name)
	if len(s[3]) > 0 {
		tag = s[3][1:]
	}
	repository = s[2]
	if len(s[1]) > 0 {
		registry = s[1][0 : len(s[1])-1]
	}
	return registry, repository, tag
}
