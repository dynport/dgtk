package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/dynport/gocloud/aws/s3"
	"github.com/dynport/gossh"
)

var logger = log.New(os.Stderr, "", 0)

const sshExample = "ubuntu@127.0.0.1"

func main() {
	dir := flag.String("dir", "", "Dir to build. Default: current directory")
	host := flag.String("host", os.Getenv("DEV_HOST"), "Host to build on. Example: "+sshExample)
	deploy := flag.String("deploy", "", "Deploy to host after building. Example: "+sshExample)
	bucket := flag.String("bucket", "", "Upload binary to s3 bucket after building")
	verbose := flag.Bool("verbose", false, "Build using -v flag")

	flag.Parse()
	logger.Printf("running with %q", *host)
	b := &build{Host: *host, Dir: *dir, DeployTo: *deploy, Bucket: *bucket, verbose: *verbose}
	e := b.Run()
	if e != nil {
		logger.Fatal("ERROR: %s", e)
	}
}

type build struct {
	Host     string
	Dir      string
	Bucket   string
	DeployTo string
	verbose  bool
}

func benchmark(message string) func() {
	started := time.Now()
	logger.Printf("started  %s", message)
	return func() {
		logger.Printf("finished %s in %.06f", message, time.Since(started).Seconds())
	}
}

func (r *build) deps() ([]string, error) {
	s, e := r.exec("go", "list", "-f", `{{ join .Deps " " }}`)
	if e != nil {
		return nil, e
	}
	return strings.Fields(s), nil
}

func (r *build) exec(cmd string, vals ...string) (string, error) {
	c := exec.Command(cmd, vals...)
	c.Dir = r.Dir
	out, e := c.CombinedOutput()
	if e != nil {
		return "", e
	}
	return string(out), nil
}

func (r *build) currentPackage() (string, error) {
	s, e := r.exec("go", "list")
	if e != nil {
		return "", e
	}
	return strings.TrimSpace(s), nil
}

func (r *build) filesMap() (map[string]os.FileInfo, error) {
	cp, e := r.currentPackage()
	if e != nil {
		return nil, e
	}
	pkgs, e := r.deps()
	if e != nil {
		return nil, e
	}
	pkgs = append(pkgs, cp)
	files := map[string]os.FileInfo{}
	sum := int64(0)
	for _, p := range pkgs {
		if !strings.Contains(p, ".") {
			continue
		}
		prefix := os.ExpandEnv("$GOPATH/src")
		dbg.Printf("walking %q", p)
		e := filepath.Walk(os.ExpandEnv(prefix+"/"+p+"/"), func(p string, info os.FileInfo, e error) error {
			skip := func() bool {
				for _, s := range []string{".git", ".bzr", ".hg"} {
					if strings.Contains(p, "/"+s+"/") {
						return true
					}
				}
				return false
			}()
			if skip {
				return nil
			}
			if _, ok := files[p]; !ok {
				sum += info.Size()
				files[p] = info
			}
			return nil
		})
		if e != nil {
			return nil, e
		}
	}
	return files, nil
}

func (b *build) createArchive() (string, error) {
	defer benchmark("create archive")()
	var name string
	e := func() error {
		f, e := ioutil.TempFile("/tmp", "gobuild-archive-")
		if e != nil {
			return e
		}
		name = f.Name()
		defer f.Close()
		files, e := b.filesMap()
		if e != nil {
			return e
		}

		gz := gzip.NewWriter(f)
		sum := int64(0)
		defer gz.Close()
		t := tar.NewWriter(gz)
		defer t.Close()

		for p, info := range files {
			name := strings.TrimPrefix(p, os.ExpandEnv("$GOPATH/src/"))
			if info.IsDir() {
				continue
			}
			dbg.Printf("adding %q", p)
			h := &tar.Header{ModTime: info.ModTime(), Size: info.Size(), Mode: int64(info.Mode()), Name: name}
			e = t.WriteHeader(h)
			if e != nil {
				return e
			}
			e := func() error {
				f, e := os.Open(p)
				if e != nil {
					return e
				}
				defer f.Close()
				i, e := io.Copy(t, f)
				sum += i
				return e
			}()
			if e != nil {
				return e
			}
		}
		dbg.Printf("%s", sizePretty(sum))
		return nil
	}()
	return name, e
}

type buildConfig struct {
	Current string
	Sudo    bool
	Verbose bool
	Version string
}

func (b *buildConfig) Goroot() string {
	return "{{ .BuildHome }}/.go/go-{{ .Version }}/go"
}

func (b *buildConfig) Gopath() string {
	return "{{ .BuildHome }}/{{ .Current }}"
}

func (b *buildConfig) BuildHome() string {
	return "$HOME/.gobuild"
}

func (b *buildConfig) BinName() string {
	return path.Base(b.Current)
}

func (b *build) Run() error {
	defer benchmark("build")()
	currentPkg, e := b.currentPackage()
	if e != nil {
		return e
	}

	cfg, e := parseConfig(b.Host)
	if e != nil {
		return e
	}
	dbg.Printf("using config %#v", cfg)
	con, e := cfg.Connection()
	if e != nil {
		return e
	}
	defer con.Close()

	name, e := b.createArchive()
	if e != nil {
		return e
	}
	dbg.Printf("created archive at %q", name)
	defer os.RemoveAll(name)
	f, e := os.Open(name)
	if e != nil {
		return e
	}
	defer f.Close()

	ses, e := con.NewSession()
	if e != nil {
		return e
	}
	ses.Stdin = f
	ses.Stdout = os.Stdout
	ses.Stderr = os.Stderr

	buildCfg := &buildConfig{
		Current: currentPkg,
		Sudo:    cfg.User != "root",
		Verbose: b.verbose,
		Version: "1.3.1",
	}

	cmd := renderRecursive(buildCmd, buildCfg)
	e = ses.Run(cmd)
	if e != nil {
		return e
	}

	name = path.Base(currentPkg)
	var binPath string

	if b.Bucket != "" || b.DeployTo != "" {
		defer os.RemoveAll(binPath)
		e = func() error {
			ses, e := con.NewSession()
			if e != nil {
				return e
			}
			defer ses.Close()

			f, e := ioutil.TempFile("/tmp", "gobuild-bin-")
			if e != nil {
				return e
			}
			binPath = f.Name()
			defer f.Close()
			ses.Stdout = f
			ses.Stderr = os.Stderr

			cmd := renderRecursive("cat {{ .Gopath }}/bin/{{ .BinName }}", buildCfg)
			e = ses.Run(cmd)
			if e != nil {
				return e
			}
			return nil
		}()
		if e != nil {
			return e
		}
	}
	if name != "" {
		defer os.RemoveAll(binPath)
	}
	if b.Bucket != "" {
		e = func() error {
			f, e := os.Open(binPath)
			if e != nil {
				return e
			}
			defer f.Close()
			client := s3.NewFromEnv()
			client.CustomEndpointHost = "s3-eu-west-1.amazonaws.com"

			parts := strings.Split(b.Bucket, "/")
			bucket := parts[0]
			key := name
			if len(parts) > 1 {
				key = strings.Join(parts[1:], "/") + "/" + key
			}
			logger.Printf("uploading to bucket=%q key=%q", bucket, key)
			return client.PutStream(bucket, key, f, nil)
		}()
		if e != nil {
			return e
		}
		logger.Printf("uploaded to bucket %q", b.Bucket)
	}

	if b.DeployTo != "" {
		e := func() error {
			cfg, e := parseConfig(b.DeployTo)
			if e != nil {
				return e
			}
			con, e := cfg.Connection()
			if e != nil {
				return e
			}
			defer con.Close()
			ses, e := con.NewSession()
			if e != nil {
				return e
			}
			defer ses.Close()
			f, e := os.Open(binPath)
			if e != nil {
				return e
			}
			defer f.Close()
			ses.Stdin = f
			ses.Stdout = os.Stdout
			ses.Stderr = os.Stderr
			s := struct {
				Name string
				Sudo bool
			}{
				Name: name, Sudo: cfg.User != "root",
			}
			cmd := renderRecursive("cd /usr/local/bin && {{ if .Sudo }}sudo {{ end }}cat - > {{ .Name }}.tmp && {{ if .Sudo }}sudo {{ end }}chmod 0755 {{ .Name }}.tmp && {{ if .Sudo }}sudo {{ end }}mv {{ .Name }}.tmp {{ .Name }}", s)
			dbg.Printf("%s", cmd)
			return ses.Run(cmd)
		}()
		if e != nil {
			return e
		}
	}
	return nil
}

func parseConfig(s string) (*gossh.Config, error) {
	cfg := &gossh.Config{}
	parts := strings.Split(s, "@")
	switch len(parts) {
	case 0:
		return nil, fmt.Errorf("Host must be set")
	case 1:
		cfg.Host = parts[0]
	case 2:
		cfg.User = parts[0]
		cfg.Host = parts[1]
	case 3:
		return nil, fmt.Errorf("format of host %q not understood", s)
	}
	return cfg, nil
}

const buildCmd = `#!/bin/bash
export BUILD_HOME={{ .BuildHome }}
export GOPATH={{ .Gopath }}
export GOROOT={{ .Goroot }}
export PATH=$GOROOT/bin:$PATH

if [ ! -f $GOROOT/bin/go ]; then
  echo "installing go {{ .Version }}"
  tmp=$(dirname $GOROOT)
  mkdir -p $tmp
  cd $tmp
  curl -sL "https://storage.googleapis.com/golang/go{{ .Version }}.linux-amd64.tar.gz" | tar xfz -
fi

set -xe
rm -Rf $GOPATH
mkdir -p $GOPATH/src
cd $GOPATH/src
tar xfz -
cd {{ .Current }}
go get {{ if .Verbose }}-v{{ end }} .
{{ with .Sudo }}sudo {{ end }}cp $GOPATH/bin/{{ .BinName }} /usr/local/bin/
`

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)

func renderRecursive(tpl string, i interface{}) string {
	s := tpl
	for j := 0; j < 10; j++ {
		rendered := mustRender([]byte(s), i)
		if rendered == s {
			return rendered
		}
		s = rendered
	}
	logger.Fatal("rendering loop, rendered 10 times")
	return ""
}

func mustRender(raw []byte, i interface{}) string {
	out, e := render(raw, i)
	if e != nil {
		logger.Fatal(e)
	}
	return out
}

func render(raw []byte, i interface{}) (string, error) {
	tpl, e := template.New(string(raw)).Parse(string(raw))
	if e != nil {
		return "", e
	}
	buf := &bytes.Buffer{}
	e = tpl.Execute(buf, i)
	if e != nil {
		return "", e
	}
	return buf.String(), nil
}

var (
	oneKb = 1024.0
	oneMb = oneKb * 1024.0
	oneGb = oneMb * 1024.0
)

func sizePretty(raw int64) string {
	f := float64(raw)
	if f < oneKb {
		return fmt.Sprintf("%.0f", f)
	} else if f < oneMb {
		return fmt.Sprintf("%.2fKB", f/oneKb)
	} else if f < oneGb {
		return fmt.Sprintf("%.2fMB", f/oneMb)
	} else {
		return fmt.Sprintf("%.2fGB", f/oneGb)
	}
}
