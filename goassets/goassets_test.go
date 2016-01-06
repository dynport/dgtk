package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setup() (string, error) {
	root, err := ioutil.TempDir("/tmp", "goassets-")
	if err != nil {
		return "", err
	}
	d := filepath.Join(root, "tmp")
	err = func() error {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
		f, e := os.Create(filepath.Join(d, "a.txt"))
		if e != nil {
			return e
		}
		defer f.Close()
		_, err := f.Write([]byte("just a test"))
		return err
	}()
	if err != nil {
		os.RemoveAll(root)
		return "", err
	}
	return root, nil
}

func TestDefaults(t *testing.T) {
	d, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	c := exec.Command("go", "run", "goassets.go", "--file", filepath.Join(d, "tmp", "assets.go"), d)
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s", err, string(b))
	}
	b, err = ioutil.ReadFile(filepath.Join(d, "tmp", "assets.go"))
	if err != nil {
		t.Fatal(err)
	}
	bs := string(b)
	if v, ex := string(bs), "package tmp"; !strings.Contains(v, ex) {
		t.Errorf("expected %q to contain %q", v, ex)
	}
}

func TestGoassets(t *testing.T) {
	d, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	c := exec.Command("go", "run", "goassets.go", "--file", filepath.Join(d, "tmp", "assets.go"), "--pkg", "main", filepath.Join(d, "tmp"))
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s", err, string(b))
	}
	b, err = ioutil.ReadFile(filepath.Join(d, "tmp", "assets.go"))
	if err != nil {
		t.Fatal(err)
	}
	if v, ex := string(b), "package main"; !strings.HasPrefix(v, ex) {
		t.Errorf("expected %q to start with %q", v, ex)
	}

	if err := ioutil.WriteFile(filepath.Join(d, "tmp", "main.go"), []byte(testProgram), 0644); err != nil {
		t.Fatal(err)
	}

	c = exec.Command("go", "run", "main.go", "assets.go")
	c.Dir = filepath.Join(d, "tmp")
	b, err = c.CombinedOutput()
	s := string(b)
	if err != nil {
		t.Fatalf("%s: %s", err, s)
	}

	for _, i := range []string{`a.txt: "just a test"`, `name: "a.txt"`} {
		if !strings.Contains(s, i) {
			t.Errorf("expected %q to contain %q", s, i)
		}
	}
}

const testProgram = `package main

import (
	"fmt"
	"log"
	"os"
	"io/ioutil"
)

var logger = log.New(os.Stdout, "", 0)

func main() {
	f, e := FileSystem("").Open("a.txt")
	if e != nil {
		logger.Fatal(e)
	}
	stat, e := f.Stat()
	if e != nil {
		logger.Fatal(e)
	}
	fmt.Printf("name: %q\n", stat.Name())
	b, e := ioutil.ReadAll(f)
	if e != nil {
		logger.Fatal(e)
	}
	fmt.Printf("a.txt: %q", string(b))
}

`
