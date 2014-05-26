package main

import (
	"time"

	"github.com/dynport/dgtk/vmware"
)

type Clone struct {
	VmName       string `cli:"type=arg required=true"`
	SnapshotName string `cli:"type=arg"`
}

func (action *Clone) Run() error {
	logger.Printf("running with name=%q and snapshot=%q", action.VmName, action.SnapshotName)
	vms, e := vmware.AllWithTemplates()
	if e != nil {
		return e
	}
	vm := vms.FindFirst(action.VmName)
	clone, e := vmware.Create(vm, action.SnapshotName)
	if e != nil {
		return e
	}

	started := time.Now()
	e = clone.Start()
	logger.Printf("started in %.3f", time.Since(started).Seconds())
	return e
}
