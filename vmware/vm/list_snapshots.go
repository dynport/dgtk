package main

import (
	"github.com/dynport/dgtk/vmware"
	"log"
)

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
		log.Println(sn.Name)
	}
	return nil
}
