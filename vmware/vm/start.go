package main

import (
	"github.com/dynport/dgtk/vmware"
)

type StartAction struct {
	Name string `cli:"type=arg required=false"`
	Gui  bool   `cli:"type=flag long=gui"`
}

func (action *StartAction) Run() error {
	vm, e := findFirst(action.Name)
	if e != nil {
		return e
	}
	return vmware.Start(vm.Path, action.Gui)
}

func init() {
	router.Register("start", &StartAction{})
}
