package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/dynport/dgtk/dockerclient"
	"github.com/dynport/dgtk/git"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Build struct {
	GitRepository   string `yaml:"repository"`
	Tag             string `yaml:"tag"`
	Proxy           string `yaml:"proxy"`
	Root            string
	DockerHost      string
	dockerfileAdded bool
}

func (b *Build) Build() (string, error) {
	buf := &bytes.Buffer{}
	e := b.buildArchive(buf)
	if e != nil {
		return "", e
	}
	client := dockerclient.New(b.DockerHost, 4243)
	return client.Build(buf, &dockerclient.BuildImageOptions{Tag: b.Tag, Callback: callback})
}

func (b *Build) buildArchive(w io.Writer) error {
	t := tar.NewWriter(w)
	defer t.Flush()
	defer t.Close()
	if b.GitRepository != "" {
		repo := &git.Repository{Origin: b.GitRepository}
		e := repo.Init()
		if e != nil {
			return e
		}
		if e := repo.Tar(t); e != nil {
			return e
		}
	}
	if e := b.addFilesToArchive(b.Root, t); e != nil {
		return e
	}
	if !b.dockerfileAdded {
		return fmt.Errorf("archive must contain a Dockerfile")
	}
	return nil
}

func (build *Build) addFilesToArchive(root string, t *tar.Writer) error {
	return filepath.Walk(root, func(p string, info os.FileInfo, e error) error {
		if e == nil && p != root {
			var e error
			name := strings.TrimPrefix(p, root+"/")
			header := &tar.Header{Name: name, ModTime: info.ModTime().UTC()}
			if info.IsDir() {
				header.Typeflag = tar.TypeDir
				header.Mode = 0755
				e = t.WriteHeader(header)
			} else {
				header.Mode = 0644
				b, e := ioutil.ReadFile(p)
				if e != nil {
					return e
				}
				if name == "Dockerfile" {
					build.dockerfileAdded = true
					if build.Proxy != "" {
						df := NewDockerfile(b)
						b = df.MixinProxy(build.Proxy)
					}
				}
				header.Size = int64(len(b))
				e = t.WriteHeader(header)
				if e != nil {
					return e
				}
				_, e = t.Write(b)
				if e != nil {
					return e
				}
			}
			if e != nil {
				return e
			}
		}
		return nil
	})
}
