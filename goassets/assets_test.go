package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func cleanup(t *testing.T) {
	os.Remove(modulePath)
	os.RemoveAll("./tmp")
}

var modulePath = "./fixtures/assets.go"

func TestIntegration(t *testing.T) {
	cleanup(t)
	os.MkdirAll("./tmp", 0755)
	b, err := ioutil.ReadFile("goassets-test/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("./tmp/main.go", b, 0755); err != nil {
		t.Fatal(err)
	}
	if fileExists("./tmp/assets.go") {
		t.Fatal("expected file to not exist")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", filepath.Join(wd, "goassets.go"), filepath.Join(wd, "fixtures"))
	cmd.Dir = "tmp"
	b, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s", string(b), err)
	}

	cmd = exec.Command("go", "run", "main.go", "assets.go")
	cmd.Dir = "tmp"
	b, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s", err, string(b))
	}
	out := string(b)
	if !fileExists("./tmp/assets.go") {
		t.Error("expected file to exist")
	}

	for _, i := range []string{"a.html: 21", "vendor/jquery.js: 15"} {
		if !strings.Contains(out, i) {
			t.Errorf("expected %q to contain %q", out, i)
		}
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
