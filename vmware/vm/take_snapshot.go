package main

import (
	"github.com/dynport/dgtk/vmware"
)

type TakeSnapshotsAction struct {
	VmName       string `cli:"type=arg required=true"`
	SnapshotName string `cli:"type=arg required=true"`
}

func (action *TakeSnapshotsAction) Run() error {
	vm, e := findFirst(action.VmName)
	if e != nil {
		return e
	}
	return vmware.TakeSnapshot(vm.Path, action.SnapshotName)
}

func init() {
	router.Register("snapshots/take", &ListSnapshotsAction{})
}
