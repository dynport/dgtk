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
	Memory       int    `cli:"opt --memory default=1024"`
	Cpus         int    `cli:"opt --cpus default=1"`
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
	logger.Printf("using %d cpus", action.Cpus)
	e = clone.ModifyCpu(action.Cpus)
	if e != nil {
		return e
	}
	logger.Printf("using %d memory", action.Memory)
	e = clone.ModifyMemory(action.Memory)
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
