package dockerclient

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/dockerclient/docker"
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
	return imageDetails, dh.getJSON(dh.url()+"/images/"+id+"/json", imageDetails)
}

// Get the given image's history.
func (dh *DockerHost) ImageHistory(id string) (imageHistory *docker.ImageHistory, e error) {
	imageHistory = &docker.ImageHistory{}
	e = dh.getJSON(dh.url()+"/images/"+id+"/history", imageHistory)
	return imageHistory, e
}

type BuildImageOptions struct {
	Tag     string
	Quite   bool
	NoCache bool

	Callback func(s *JSONMessage)
}

func (opts *BuildImageOptions) encode() string {
	v := url.Values{}
	if opts != nil {
		if opts.Tag != "" {
			v.Add("t", opts.Tag)
		}
		if opts.NoCache {
			v.Add("nocache", "true")
		}
		if opts.Quite {
			v.Add("quiet", "true")
		}
	}
	return v.Encode()
}

// Create a new image from the given dockerfile. If name is non empty the new image is named accordingly. If a writer is
// given it is used to send the docker output to.
func (dh *DockerHost) BuildDockerfile(dockerfile string, opts *BuildImageOptions) (imageId string, e error) {
	r, e := dh.createDockerfileArchive(dockerfile)
	if e != nil {
		return "", e
	}
	return dh.Build(r, opts)

}

// Build a container image from a tar or tar.gz Reader
func (dh *DockerHost) Build(r io.Reader, opts *BuildImageOptions) (imageId string, e error) {
	u := dh.url() + "/build"
	if opts == nil {
		opts = &BuildImageOptions{}
	}
	if enc := opts.encode(); enc != "" {
		u += "?" + enc
	}
	rsp, e := dh.postWithContentType(u, "application/tar", r)
	if e != nil {
		return "", e
	}
	defer rsp.Body.Close()

	ib := &imageBuildResponseHandler{userHandler: opts.Callback}
	return ib.Run(rsp.Body)
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
	if e != nil {
		return e
	}
	return rsp.Body.Close()
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

	return handleJSONStream(rsp.Body, handlePullImageMessage)
}

func handlePullImageMessage(msg *JSONMessage) {
	if e := msg.Err(); e != nil {
		log.Printf("error creating image: %s", e)
	}

	if msg.Status == "Download complete" {
		log.Printf("finished downloading %s", msg.Id)
	}
}

type PushImageOptions struct {
	Callback func(s *JSONMessage)
}

// Push the given image to the registry. The name should be <registry>/<repository>.
func (dh *DockerHost) PushImage(name string, opts *PushImageOptions) error {
	if name == "" {
		return fmt.Errorf("no image name given")
	}

	if opts == nil {
		opts = &PushImageOptions{Callback: handlePullImageMessage}
	}

	registry, image, tag := splitImageName(name)
	if registry == "" {
		return fmt.Errorf("no registry given")
	}

	buf := &bytes.Buffer{}
	buf.WriteString(FAKE_AUTH)
	url := dh.url() + "/images/" + registry + "/" + image + "/push?tag=" + tag

	rsp, e := dh.postWithReader(url, buf)
	if e != nil {
		return e
	}
	defer rsp.Body.Close()

	return handleJSONStream(rsp.Body, opts.Callback)
}

func handlePushImageMessage(msg *JSONMessage) {
	if e := msg.Err(); e != nil {
		log.Printf("error pushing image: %s", e)
	}
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
		_, e := dh.ImageDetails(repository + ":" + tag)
		if e != nil {
			if e.Error() == "resource not found" {
				time.Sleep(1 * time.Second)
				continue
			}
			return e
		}
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
