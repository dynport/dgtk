package main

import (
	"flag"
	"github.com/dynport/dgtk/dockerbuild"
	"log"
	"path/filepath"
)

var (
	buildHost  = flag.String("H", "", "Build Host (e.g. 127.0.0.1)")
	tag        = flag.String("T", "", "Tag build with (e.g. elasticsearch)")
	proxy      = flag.String("X", "", "Http Proxy to use (e.g. http://127.0.0.1:1234)")
	repository = flag.String("R", "", "Git repository to add to docker archive (e.g. git@github.com:test/repo.git)")
)

func main() {
	flag.Parse()
	root := flag.Arg(0)
	if root == "" {
		log.Fatal("root must be provided")
	}
	root, e := filepath.Abs(root)
	if e != nil {
		log.Fatal(e.Error())
	}
	build := &dockerbuild.Build{Root: root, Tag: *tag, Proxy: *proxy, GitRepository: *repository, DockerHost: *buildHost}
	if build.DockerHost == "" {
		log.Fatal("-H must be provided")
	}
	imageId, e := build.Build()
	if e != nil {
		log.Fatal(e.Error())
	}
	log.Printf("built image id %q", imageId)
}
