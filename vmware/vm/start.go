package main

import (
	"github.com/dynport/dgtk/vmware"
)

type StartAction struct {
	Name string `cli:"type=arg required=true"`
	Gui  bool   `cli:"type=opt long=gui"`
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
