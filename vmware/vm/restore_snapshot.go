package main

import (
	"github.com/dynport/dgtk/vmware"
)

type RestoreSnapshotAction struct {
	VmName       string `cli:"type=arg required=true"`
	SnapshotName string `cli:"type=arg required=true"`
}

func (action *RestoreSnapshotAction) Run() error {
	vm, e := findFirst(action.VmName)
	if e != nil {
		return e
	}
	return vmware.RestoreSnapshot(vm.Path, action.SnapshotName)
}
