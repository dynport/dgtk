package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dynport/dgtk/dockerbuild"
)

var (
	dockerHost = flag.String("H", os.Getenv("DOCKER_BUILD_HOST"), "Docker Host (e.g. 127.0.0.1:4243)")
	tag        = flag.String("T", "", "Tag build with (e.g. elasticsearch)")
	proxy      = flag.String("X", os.Getenv("DOCKER_BUILD_PROXY"), "Http Proxy to use (e.g. http://127.0.0.1:1234)")
	repository = flag.String("R", "", "Git repository to add to docker archive (e.g. git@github.com:test/repo.git)")
)

func parseDockerHost(hostAndPort string) (string, int) {
	parts := strings.SplitN(hostAndPort, ":", 2)
	port := 4243
	if len(parts) == 2 {
		port, _ = strconv.Atoi(parts[1])
	}
	return parts[0], port
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	if root == "" {
		root = "."
	}
	root, e := filepath.Abs(root)
	if e != nil {
		log.Fatal(e.Error())
	}
	dockerHost, dockerPort := parseDockerHost(*dockerHost)
	build := &dockerbuild.Build{
		Root:           root,
		DockerImageTag: *tag,
		Proxy:          *proxy,
		GitRepository:  *repository,
		DockerHost:     dockerHost,
		DockerPort:     dockerPort,
	}
	if build.DockerHost == "" {
		log.Fatal("-H must be provided")
	}
	imageId, e := build.Build()
	if e != nil {
		log.Fatal(e.Error())
	}
	log.Printf("built image id %q", imageId)
}
