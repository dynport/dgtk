package main

import (
	"flag"
	"log"
	"path/filepath"
)

var (
	buildHost  = flag.String("H", "", "Build Host")
	tag        = flag.String("T", "", "Tag build with")
	proxy      = flag.String("X", "", "Http Proxy to use")
	repository = flag.String("R", "", "Git repository to add to docker archive")
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
	build := &Build{Root: root}
	build.LoadConfig()
	if *tag != "" {
		build.Tag = *tag
	}
	if *proxy != "" {
		build.Proxy = *proxy
	}
	if *repository != "" {
		build.GitRepository = *repository
	}
	if *buildHost == "" {
		log.Fatal("-H must be provided")
	}
	build.DockerHost = *buildHost
	imageId, e := build.Build()
	if e != nil {
		log.Fatal(e.Error())
	}
	log.Printf("built image id %q", imageId)
}
