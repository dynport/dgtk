package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/dynport/dgtk/goassets/goassets"
	. "github.com/smartystreets/goconvey/convey"
)

func cleanup(t *testing.T) {
	os.Remove(modulePath)
	os.RemoveAll("./tmp")
}

var modulePath = "./fixtures/assets.go"

func TestAssets(t *testing.T) {
	assets := &goassets.Assets{Paths: []string{"./fixtures"}, CustomPackagePath: modulePath}
	Convey("Assets", t, func() {
		Convey("AssetsPaths", func() {
			So(assets, ShouldNotBeNil)
			paths, e := assets.AssetPaths()
			So(e, ShouldBeNil)
			So(len(paths), ShouldEqual, 2)
			So(paths[0].Path, ShouldEqual, "fixtures/a.html")
			So(paths[0].Key, ShouldEqual, "a.html")
			So(paths[1].Path, ShouldEqual, "fixtures/vendor/jquery.js")
			So(paths[1].Key, ShouldEqual, "vendor/jquery.js")
		})
		Convey("Build", func() {
			Convey("raise an error when file exists", func() {
				e := ioutil.WriteFile(modulePath, []byte("//just some comment"), 0644)
				if e != nil {
					t.Fatal(e.Error())
				}
				e = assets.Build()
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldContainSubstring, "already exists")
			})

			Convey("with assets.go not exising", func() {
				cleanup(t)
				e := assets.Build()
				So(e, ShouldBeNil)
			})
		})
	})
}

func fileExists(p string) bool {
	_, e := os.Stat(p)
	if e != nil {
		return false
	}
	return true
}

func TestIntegration(t *testing.T) {
	Convey("Integration test", t, func() {
		cleanup(t)
		os.MkdirAll("./tmp", 0755)
		e := ioutil.WriteFile("./tmp/main.go", []byte(TEST_FILE), 0755)
		if e != nil {
			t.Fatal(e.Error())
		}
		So(fileExists("./tmp/assets.go"), ShouldBeFalse)
		cmd := exec.Command("bash", "-c", "cd tmp && goassets ../fixtures && go run *.go")
		b, e := cmd.CombinedOutput()
		So(e, ShouldBeNil)
		out := string(b)
		So(fileExists("./tmp/assets.go"), ShouldBeTrue)
		So(out, ShouldContainSubstring, "a.html: 21")
		So(out, ShouldContainSubstring, "vendor/jquery.js: 15")
	})
}

const TEST_FILE = `
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("running")
	for _, name := range assetNames() {
		b, e := readAsset(name)
		if e != nil {
			fmt.Println("ERROR: " + e.Error())
			os.Exit(1)
		}
		fmt.Printf("%v: %d\n", name, len(b))
	}
}
`
