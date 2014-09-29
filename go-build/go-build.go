package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/dynport/gocloud/aws/s3"
)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)
var logger = log.New(os.Stderr, "", 0)

func main() {
	if e := run(); e != nil {
		log.Fatal(e)
	}
}

func run() error {
	var dir = flag.String("d", ".", "Path to project")
	var s3Bucket = flag.String("bucket", "", "S3 Bucket to upload binary to")
	flag.Parse()
	p, e := build(*dir)
	if *s3Bucket != "" {
		client := s3.NewFromEnv()
		e := func() error {
			f, e := os.Open(p)
			if e != nil {
				return e
			}
			defer f.Close()
			log.Printf("uploading %q to %q", p, *s3Bucket)
			client.PutStream("bucket", path.Base(p), f, nil)
			return nil
		}()
		if e != nil {
			return e
		}
	}
	return e
}

func splitBucket(bucketWithPrefix string) (string, string) {
	parts := strings.SplitN(bucketWithPrefix, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return bucketWithPrefix, ""
}

func build(dir string) (string, error) {
	buf := &bytes.Buffer{}
	c := exec.Command("go", "list", "-f", "{{ .Deps }}")
	c.Dir = dir
	fmt.Printf("building in %q\n", c.Dir)
	c.Stderr = os.Stderr
	c.Stdout = buf
	e := c.Run()
	if e != nil {
		return "", e
	}
	repos := map[string]struct{}{}
	for _, d := range strings.Fields(buf.String()) {
		if strings.Contains(d, ".") {
			if repo := findRepo(os.ExpandEnv("$GOPATH/src/") + d); repo != "" {
				repos[repo] = struct{}{}
			}
		}
	}

	name, e := executeIn(dir, "go", "list")
	if e != nil {
		return "", e
	}
	abs, e := filepath.Abs(dir)
	if e != nil {
		return "", e
	}

	repo := findRepo(abs)
	if repo == "" {
		return "", fmt.Errorf("not a git repository")
	}

	versions, e := gitHistory(repo)
	if e != nil {
		return "", e
	}
	status := &BuildStatus{Name: strings.TrimSpace(name), Versions: versions}
	status.Changes, e = gitChanges(dir)
	if e != nil {
		return "", e
	}
	for r := range repos {
		versions, e := gitHistory(r)
		if e != nil {
			return "", e
		}
		name := strings.TrimPrefix(r, os.ExpandEnv("$GOPATH/src/"))
		s := &BuildStatus{Name: name, Versions: versions}
		s.Changes, e = gitChanges(r)
		if e != nil {
			return "", e
		}
		status.Dependencies = append(status.Dependencies, s)
	}
	b, e := json.MarshalIndent(status, "", "  ")
	if e != nil {
		return "", e
	}
	binPath := "/tmp/" + path.Base(dir)
	encoded := base64.StdEncoding.EncodeToString(b)
	_, e = executeIn(dir, "go", "build", "-o", binPath, "-ldflags", "-X main.BUILD_INFO "+encoded)
	fmt.Printf("built binary %q\n", binPath)
	return binPath, nil
}

func findRepo(start string) string {
	parts := strings.Split(start, "/")
	for i := 0; i < len(parts)-1; i++ {
		p := strings.Join(parts[0:len(parts)-i], "/")
		_, e := os.Stat(p + "/.git")
		if e == nil {
			return p
		}
	}
	return ""
}

func executeIn(dir, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = buf
	c.Dir = dir
	e := c.Run()
	return buf.String(), e
}

func gitHistory(dir string) ([]string, error) {
	out, e := executeIn(dir, "git", "--git-dir="+dir+"/.git", "log", `--pretty=%h %ai`, "-n", "10")
	if e != nil {
		return nil, e
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

type BuildStatus struct {
	Name         string
	Versions     []string
	Changes      bool
	Dependencies []*BuildStatus
}

func gitChanges(dir string) (bool, error) {
	out, e := executeIn(dir, "git", "status", "--porcelain", ".")
	if e != nil {
		return false, e
	}
	return len(strings.TrimSpace(out)) > 0, nil
}
