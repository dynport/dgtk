package main

import (
	"fmt"
	"time"

	"github.com/dynport/dgtk/vmware"
)

type Clone struct {
	VmName       string `cli:"type=arg required=true"`
	SnapshotName string `cli:"type=arg"`
	Name         string `cli:"opt --name"`
}

func (action *Clone) Run() error {
	out := fmt.Sprintf("cloning vm %q", action.VmName)
	if action.SnapshotName != "" {
		out += fmt.Sprintf(" and snapshot=%q", action.SnapshotName)
	}
	logger.Printf(out)
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
	if e != nil {
		return e
	}

	if action.Name != "" {
		return vmware.UpdateTag(&vmware.Tag{VmId: clone.Id(), Key: "Name", Value: action.Name})
	}
	return nil
}
