package main

import "github.com/dynport/dgtk/vmware"

type ListSnapshotsAction struct {
	Name string `cli:"type=arg required=true"`
}

func (action *ListSnapshotsAction) Run() error {
	vm, e := findFirst(action.Name)
	if e != nil {
		return e
	}
	snapshots, e := vmware.ListSnapshots(vm.Path)
	if e != nil {
		return e
	}
	for _, sn := range snapshots {
		logger.Println(sn.Name)
	}
	return nil
}
