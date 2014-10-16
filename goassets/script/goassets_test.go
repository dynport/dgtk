package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func setup(t *testing.T) {
	os.RemoveAll("tmp")
	So(os.MkdirAll("tmp", 0755), ShouldBeNil)
	e := func() error {
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

func TestGoassets(t *testing.T) {
	Convey("Integration", t, func() {
		Convey("running with defaults", func() {
			setup(t)
			c := exec.Command("go", "run", "goassets.go", "--file", "tmp/assets.go", "tmp")
			b, e := c.CombinedOutput()
			if e != nil {
				t.Fatal(e.Error() + ": " + string(b))
			}
			b, e = ioutil.ReadFile("tmp/assets.go")
			So(e, ShouldBeNil)
			bs := string(b)
			So(bs, ShouldStartWith, "package tmp")
			So(e, ShouldBeNil)

		})

		Convey("Running with custom package", func() {
			setup(t)
			c := exec.Command("go", "run", "goassets.go", "--file", "tmp/assets.go", "--pkg", "main", "tmp")
			b, e := c.CombinedOutput()
			if e != nil {
				t.Fatal(e.Error() + ": " + string(b))
			}
			b, e = ioutil.ReadFile("tmp/assets.go")
			So(e, ShouldBeNil)
			bs := string(b)
			So(bs, ShouldStartWith, "package main")
			So(e, ShouldBeNil)

			e = ioutil.WriteFile("tmp/main.go", []byte(testProgram), 0644)
			So(e, ShouldBeNil)

			wd, e := os.Getwd()
			if e != nil {
				t.Fatal(e)
			}
			defer os.Chdir(wd)

			So(os.Chdir("tmp"), ShouldBeNil)

			b, e = exec.Command("go", "run", "main.go", "assets.go").CombinedOutput()
			if e != nil {
				t.Fatal(e.Error() + ": " + string(b))
			}
			So(b, ShouldNotBeNil)
			So(string(b), ShouldContainSubstring, `a.txt: "just a test"`)
			So(string(b), ShouldContainSubstring, `name: "a.txt"`)
		})
	})
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
