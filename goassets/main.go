package main

import (
	"fmt"
	"github.com/dynport/dgtk/cli"
	"os/exec"
	"path/filepath"
	"strings"
)

type action struct {
	TargetFile string   `cli:"type=opt short=t long=target default=assets.go desc='The name of the file created.'"`
	AssetPaths []string `cli:"type=arg required=true desc='Paths where raw assets are located.'"`
}

func (a *action) Run() error {
	if filepath.Dir(a.TargetFile) != "." {
		return fmt.Errorf("The target %q must be located in the directory goassets is called in.", a.TargetFile)
	}

	packageName := determinePackageByPath()

	assets := &Assets{
		Package:           packageName,
		CustomPackagePath: a.TargetFile,
		Paths:             a.AssetPaths,
	}

	if e := assets.Build(); e != nil {
		return e
	}

	return nil
}

const BYTE_LENGTH = 12

func makeLineBuffer() []string {
	return make([]string, 0, BYTE_LENGTH)
}

func determinePackageByPath() string {
	result, e := exec.Command("go", "list", "-f", "{{ .Name }}").CombinedOutput()
	if e != nil {
		logger.Fatal(string(result), e.Error())
	}
	return strings.TrimSpace(string(result))
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
