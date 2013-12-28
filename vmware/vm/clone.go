package main

import (
	"github.com/dynport/dgtk/vmware"
	"log"
	"time"
)

type Clone struct {
	VmName       string `cli:"type=arg required=true"`
	SnapshotName string `cli:"type=arg"`
}

func (action *Clone) Run() error {
	log.Printf("running with name=%q and snapshot=%q", action.VmName, action.SnapshotName)
	vm, e := findFirst(action.VmName)
	if e != nil {
		return e
	}

	clone, e := vmware.Create(vm, action.SnapshotName)
	if e != nil {
		return e
	}

	started := time.Now()
	e = clone.Start()
	log.Printf("started in %.3f", time.Since(started).Seconds())
	return e
}

func init() {
	router.Register("clone", &Clone{})
}
