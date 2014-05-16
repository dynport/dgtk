package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/dynport/dgtk/cli"
	"github.com/dynport/dgtk/goassets/goassets"
)

type action struct {
	TargetFile string   `cli:"type=opt short=t long=target default=assets.go desc='The name of the file created.'"`
	AssetPaths []string `cli:"type=arg required=true desc='Paths where raw assets are located.'"`
}

func (a *action) Run() error {
	if filepath.Dir(a.TargetFile) != "." {
		return fmt.Errorf("The target %q must be located in the directory goassets is called in.", a.TargetFile)
	}

	packageName, e := determinePackageByPath()
	if e != nil {
		logger.Fatal(e)
	}

	assets := &goassets.Assets{
		Package:           packageName,
		CustomPackagePath: a.TargetFile,
		Paths:             a.AssetPaths,
	}

	if e := assets.Build(); e != nil {
		return e
	}

	return nil
}

func determinePackageByPath() (string, error) {
	result, e := exec.Command("go", "list", "-f", "{{ .Name }}").CombinedOutput()
	if e != nil {
		wd, e2 := os.Getwd()
		if e2 != nil {
			return "", e2
		}
		return path.Base(wd), nil
	}
	return strings.TrimSpace(string(result)), nil
}

func main() {
	if e := cli.RunActionWithArgs(&action{}); e != nil {
		switch e {
		case nil, cli.ErrorHelpRequested, cli.ErrorNoRoute:
			// ignore
		default:
			logger.Fatal(e.Error())
		}
	}
}
