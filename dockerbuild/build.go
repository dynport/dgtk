package dockerbuild

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dynport/dgtk/dockerclient"
	"github.com/dynport/dgtk/git"
)

type Build struct {
	GitRepository string
	Tag           string
	Proxy         string
	Root          string

	// If this is a ruby project then add the Gemfiles to the archive separately. That way bundler's inefficiency can be
	// mitigated using docker's caching strategy. Just call copy the Gemfile's somewhere (using the 'ADD' command) and
	// run bundler on them. Then extract the sources and use the app. This way only changes to the Gemfiles will result
	// in a rerun of bundler.
	RubyProject bool

	DockerHost         string // IP of the host running docker.
	DockerPort         int    // Port docker is listening on.
	DockerHostUser     string // If set an SSH tunnel will be setup and used for communication.
	DockerHostPassword string // Password of the user, if required for SSH (public key authentication should be preferred).

	Revision        string
	dockerfileAdded bool
}

type progress struct {
	total   int64
	current int64
	started time.Time
}

func newProgress(total int64) *progress {
	return &progress{started: time.Now(), total: total}
}

func (p *progress) Write(b []byte) (int, error) {
	i := len(b)
	p.current += int64(i)
	fmt.Printf("\rupload progress %.1f%%", 100.0*float64(p.current)/float64(p.total))
	if p.current == p.total {
		fmt.Printf("\nuploaded total_size=%.3fMB in total_time%.3fs\n", float64(p.total)/(1024.0*1024.0), time.Since(p.started).Seconds())
	}
	return i, nil
}

func (b *Build) connectToDockerHost() (*dockerclient.DockerHost, error) {
	if b.DockerHostUser == "" {
		port := b.DockerPort
		if port == 0 {
			port = 4243
		}
		return dockerclient.New(b.DockerHost, port), nil
	}
	return dockerclient.NewViaTunnel(b.DockerHost, b.DockerHostUser, b.DockerHostPassword)
}

func (b *Build) Build() (string, error) {
	f, e := b.buildArchive()
	if e != nil {
		return "", e
	}
	defer func() { os.Remove(f.Name()) }()
	log.Printf("wrote file %s", f.Name())

	client, e := b.connectToDockerHost()
	if e != nil {
		return "", e
	}

	f, e = os.Open(f.Name())
	if e != nil {
		return "", e
	}
	stat, e := f.Stat()
	if e != nil {
		return "", e
	}
	progress := newProgress(stat.Size())

	r := io.TeeReader(f, progress)
	imageId, e := client.Build(r, &dockerclient.BuildImageOptions{Tag: b.Tag, Callback: callback})
	if e != nil {
		return imageId, e
	}
	if b.Tag != "" {
		imageDetails, e := client.ImageDetails(imageId)
		if e == nil {
			created, e := imageDetails.CreatedAt()
			if e != nil {
				log.Print("ERROR: " + e.Error())
			} else {
				tag := created.UTC().Format("2006-01-02T150405")
				e := client.TagImage(imageId, b.Tag, tag)
				if e != nil {
					log.Print("ERROR: " + e.Error())
				}
			}
		}
	}
	return imageId, e
}

func (b *Build) buildArchive() (*os.File, error) {
	f, e := ioutil.TempFile("/tmp", "docker_build")
	if e != nil {
		return nil, e
	}
	defer f.Close()
	t := tar.NewWriter(f)
	defer t.Flush()
	defer t.Close()
	if b.GitRepository != "" {
		repo := &git.Repository{Origin: b.GitRepository}
		e := repo.Init()
		if e != nil {
			return nil, e
		}
		e = repo.Fetch()
		if e != nil {
			return nil, e
		}
		if e := repo.WriteArchiveToTar(b.Revision, t); e != nil {
			return nil, e
		}
		if b.RubyProject {
			if e := repo.WriteFilesToTar(b.Revision, t, "Gemfile", "Gemfile.lock"); e != nil {
				return nil, e
			}
		}
	}
	if e := b.addFilesToArchive(b.Root, t); e != nil {
		return nil, e
	}
	if !b.dockerfileAdded {
		return nil, fmt.Errorf("archive must contain a Dockerfile")
	}
	return f, nil
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
