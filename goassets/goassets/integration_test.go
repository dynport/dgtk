package goassets

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIntegration(t *testing.T) {
	Convey("Integration test", t, func() {
		assets := &Assets{}
		assets.Paths = []string{
			"fixtures",
		}
		os.MkdirAll("tmp", 0755)
		dir, e := ioutil.TempDir("tmp", "assets")
		if e != nil {
			t.Fatal(e)
		}
		defer func() {
			os.RemoveAll(dir)
		}()
		assets.CustomPackagePath = dir + "/assets.go"
		e = assets.Build()
		So(e, ShouldBeNil)

		e = ioutil.WriteFile(dir+"/main.go", []byte(mainTpl), 0644)
		if e != nil {
			t.Fatal(e)
		}

		b, e := exec.Command("bash", "-c", "cd "+dir+" && go run *.go").CombinedOutput()
		if e != nil {
			t.Log("error: " + string(b))
			t.Fatal(e)
		}

		var m map[string]interface{}

		So(json.Unmarshal(b, &m), ShouldBeNil)
		So(m["Status"], ShouldEqual, "ok")
		So(m["Names"], ShouldEqual, "a.html,a.txt,b.txt,vendor/jquery.js")
	})
}

const mainTpl = `package main

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

func main() {
	names := assetNames()
	sort.Strings(names)
	m := map[string]interface{}{
		"Status": "ok",
		"Names": strings.Join(names, ","),
	}
	json.NewEncoder(os.Stdout).Encode(m)
}
`
