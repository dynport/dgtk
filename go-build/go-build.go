package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

var dbg = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)

func main() {
	if e := run(); e != nil {
		log.Fatal(e)
	}
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
			parts := strings.Split(d, "/")
			for i := 0; i < len(parts)-1; i++ {
				p := os.ExpandEnv("$GOPATH/src/" + strings.Join(parts[0:len(parts)], "/"))
				dbg.Printf("testing %q", p)
				if _, e = os.Stat(p + "/.git"); e != nil {
					continue
				}
				repos[p] = struct{}{}
			}
		}
	}

	name, e := executeIn(dir, "go", "list")
	if e != nil {
		return "", e
	}
	status := &BuildStatus{Name: strings.TrimSpace(name)}
	for r := range repos {
		out, e := executeIn(dir, "git", "--git-dir="+r+"/.git", "log", `--pretty=%h %ai`, "-n", "10")
		if e != nil {
			return "", e
		}
		name := strings.TrimPrefix(r, os.ExpandEnv("$GOPATH/src/"))
		status.Dependencies = append(status.Dependencies, &BuildStatus{Name: name, Versions: strings.Split(strings.TrimSpace(string(out)), "\n")})
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

func executeIn(dir, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = buf
	c.Dir = dir
	e := c.Run()
	return buf.String(), e
}

func run() error {
	var dir = flag.String("d", ".", "Path to project")
	flag.Parse()
	_, e := build(*dir)
	return e
}

type BuildStatus struct {
	Name         string
	Versions     []string
	Dependencies []*BuildStatus
}
