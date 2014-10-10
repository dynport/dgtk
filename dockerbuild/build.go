package dockerbuild

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dynport/dgtk/dockerclient"
	"github.com/dynport/dgtk/git"
)

type Build struct {
	Proxy string // http proxy to use for faster builds in local environment
	Root  string // root directory used for archive (should contain a dockerfile at least)

	GitRepository string // repository to check out into the docker build archive
	GitRevision   string // revision of the repository to use

	// If this is a ruby project then add the Gemfiles to the archive separately. That way bundler's inefficiency can be
	// mitigated using docker's caching strategy. Just call copy the Gemfile's somewhere (using the 'ADD' command) and
	// run bundler on them. Then extract the sources and use the app. This way only changes to the Gemfiles will result
	// in a rerun of bundler.
	RubyProject bool

	DockerHost         string // IP of the host running docker.
	DockerPort         int    // Port docker is listening on.
	DockerHostUser     string // If set an SSH tunnel will be setup and used for communication.
	DockerHostPassword string // Password of the user, if required for SSH (public key authentication should be preferred).
	DockerImageTag     string // tag used for the resulting image

	ForceBuild bool // Build even if image already exists.
	Verbose    bool

	msgList []message

	client *dockerclient.DockerHost
}

type message struct {
	Step     int
	Cached   bool
	Command  string
	Image    string
	Started  time.Time
	Finished time.Time
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

func (b *Build) connectToDockerHost() (e error) {
	if b.client != nil {
		return nil
	}

	if b.DockerHostUser == "" {
		port := b.DockerPort
		if port == 0 {
			port = 4243
		}
		b.client, e = dockerclient.New(b.DockerHost, port)
		return e
	}
	b.client, e = dockerclient.NewViaTunnel(b.DockerHost, b.DockerHostUser, b.DockerHostPassword)
	return e
}

func (b *Build) BuildAndPush() (string, error) {
	imageId, e := b.Build()
	if e != nil {
		return imageId, e
	}
	// build has connected so we can assume b.client is set
	return imageId, b.client.PushImage(b.DockerImageTag, &dockerclient.PushImageOptions{Callback: b.callbackForPush})
}

func (b *Build) Build() (string, error) {
	if e := b.connectToDockerHost(); e != nil {
		return "", e
	}

	if !b.ForceBuild {
		details, e := b.client.ImageDetails(b.DockerImageTag)
		switch e {
		case nil:
			log.Printf("image for %q already exists", b.DockerImageTag)
			return details.Id, nil
		case dockerclient.ErrorNotFound:
			// ignore
		default:
			return "", e
		}
	}

	f, e := b.buildArchive()
	if e != nil {
		return "", e
	}
	defer func() { os.Remove(f.Name()) }()

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

	return b.client.Build(r, &dockerclient.BuildImageOptions{
		Tag:      b.DockerImageTag,
		Callback: b.callbackForBuild,
	})
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
		if e := repo.WriteArchiveToTar(b.GitRevision, t); e != nil {
			return nil, e
		}
		if b.RubyProject {
			if e := repo.WriteFilesToTar(b.GitRevision, t, "Gemfile", "Gemfile.lock"); e != nil {
				return nil, e
			}
		}
	}
	if withDockerFile, e := b.addFilesToArchive(b.Root, t); e != nil {
		return nil, e
	} else if !withDockerFile {
		return nil, fmt.Errorf("archive must contain a Dockerfile")
	}
	return f, nil
}

func (build *Build) addFilesToArchive(root string, t *tar.Writer) (withDockerFile bool, e error) {
	return withDockerFile, filepath.Walk(root, func(p string, info os.FileInfo, e error) error {
		if e == nil && p != root {
			var e error
			name := strings.TrimPrefix(p, root+"/")
			header := &tar.Header{Name: name, ModTime: info.ModTime().UTC()}
			switch {
			case info.IsDir():
				if name == ".git" {
					return filepath.SkipDir
				}
				header.Typeflag = tar.TypeDir
				header.Mode = 0755
				e = t.WriteHeader(header)
			default:
				b, e := ioutil.ReadFile(p)
				if e != nil {
					return e
				}

				if name == "Dockerfile" {
					withDockerFile = true
					if build.Proxy != "" {
						b = NewDockerfile(b).MixinProxy(build.Proxy)
					}
				}

				header.Mode = 0644
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

func (b *Build) callbackForBuild(msg *dockerclient.JSONMessage) {
	if e := msg.Err(); e != nil {
		log.Printf("%s", e)
		return
	}
	msg.Stream = strings.TrimSpace(msg.Stream)
	switch {
	case msg.Stream == "":
		// ignore
	case b.Verbose:
		fmt.Println(msg.Stream)
	case strings.HasPrefix(msg.Stream, "Step "):
		parts := strings.SplitN(msg.Stream[5:], " : ", 2)
		if len(parts) != 2 { // ignore message
			return
		}
		step, e := strconv.Atoi(parts[0])
		if e != nil {
			return
		}
		newMessage := message{
			Step:    step,
			Command: parts[1],
			Started: time.Now(),
		}
		b.msgList = append(b.msgList, newMessage)
		fmt.Printf("%.3d %q", newMessage.Step, newMessage.Command)
	case msg.Stream == "---> Using cache":
		if len(b.msgList) > 0 {
			b.msgList[len(b.msgList)-1].Cached = true
		}
	case strings.HasPrefix(msg.Stream, "---> "):
		if len(b.msgList) > 0 && len(msg.Stream) == 17 {
			i := len(b.msgList) - 1
			b.msgList[i].Finished = time.Now()
			b.msgList[i].Image = msg.Stream[5:]
			status := "BUILT "
			if b.msgList[i].Cached {
				status = "CACHED"
			}
			duration := b.msgList[i].Finished.Sub(b.msgList[i].Started)
			fmt.Printf("\r%.3d [%s] [%s] %q took %5.2fs\n", b.msgList[i].Step, status, b.msgList[i].Image, b.msgList[i].Command, duration.Seconds())
		}
	}
}

func (b *Build) callbackForPush(msg *dockerclient.JSONMessage) {
	if e := msg.Err(); e != nil {
		log.Printf("%s", e)
		return
	}
	msg.Status = strings.TrimSpace(msg.Status)
	switch {
	case msg.Status == "":
		// ignore
	case b.Verbose:
		fmt.Printf("%s\n", msg.Status)
	case msg.Status == "Buffering to disk" && msg.Progress.Total > 0:
		fmt.Printf("\r[%s] Buffering: %4.2f %%", msg.Id, 100.0*float32(msg.Progress.Current)/float32(msg.Progress.Total))
	case msg.Status == "Pushing" && msg.Progress.Total > 0:
		fmt.Printf("\r[%s] Pushing:   %4.2f %%", msg.Id, 100.0*float32(msg.Progress.Current)/float32(msg.Progress.Total))
	case msg.Status == "Image successfully pushed":
		fmt.Printf("\r[%s] Pushing:   ... done\n", msg.Id)
	}
}
