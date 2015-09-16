package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func setup(t *testing.T) {
	os.RemoveAll("tmp")
	e := func() error {
		if err := os.MkdirAll("tmp", 0755); err != nil {
			return err
		}
		f, e := os.Create("tmp/a.txt")
		if e != nil {
			return e
		}
		defer f.Close()
		_, e = f.Write([]byte("just a test"))
		return e
	}()
	if e != nil {
		t.Fatal(e)
	}
}

func TestDefaults(t *testing.T) {
	setup(t)
	c := exec.Command("go", "run", "goassets.go", "--file", "tmp/assets.go", "tmp")
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s", err, string(b))
	}
	b, err = ioutil.ReadFile("tmp/assets.go")
	if err != nil {
		t.Fatal(err)
	}
	bs := string(b)
	if v, ex := string(bs), "package tmp"; !strings.Contains(v, ex) {
		t.Errorf("expected %q to contain %q", v, ex)
	}
}

func TestGoassets(t *testing.T) {
	setup(t)
	c := exec.Command("go", "run", "goassets.go", "--file", "tmp/assets.go", "--pkg", "main", "tmp")
	b, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s", err, string(b))
	}
	b, err = ioutil.ReadFile("tmp/assets.go")
	if err != nil {
		t.Fatal(err)
	}
	if v, ex := string(b), "package main"; !strings.HasPrefix(v, ex) {
		t.Errorf("expected %q to start with %q", v, ex)
	}

	if err := ioutil.WriteFile("tmp/main.go", []byte(testProgram), 0644); err != nil {
		t.Fatal(err)
	}

	c = exec.Command("go", "run", "main.go", "assets.go")
	c.Dir = "tmp"
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
