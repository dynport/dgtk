package main

import "github.com/dynport/dgtk/vmware"

type StopAction struct {
	Name   string `cli:"type=arg required=true"`
	Delete bool   `cli:"type=opt long=delete desc='delete vm after stopping'"`
}

func (action *StopAction) Run() error {
	vm, e := findFirst(action.Name)
	if e != nil {
		return e
	}
	logger.Printf("stopping vm at %s", vm.Path)
	e = vmware.Stop(vm.Path)
	if e != nil {
		return e
	}
	if action.Delete {
		logger.Printf("deleting vm at %s", vm.Path)
		return vmware.DeleteVM(vm.Path)
	}
	return nil
}
