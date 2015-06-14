package main

import (
	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
)

type ListAction struct {
}

func (list *ListAction) Run() error {
	vms, err := vmware.AllWithIPsAndTags()
	if err != nil {
		return err
	}
	table := gocli.NewTable()
	table.Add("Id", "Name", "Status", "Started", "Cpus", "Memory", "Mac", "Ip", "SoftPowerOff", "CleanShutdown")
	for _, vm := range vms {
		vmx, e := vm.Vmx()
		if e != nil {
			return e
		}
		started := ""
		if s, e := vm.StartedAt(); e == nil {
			started = s.Format("2006-01-02T15:04:05")
		} else {
			logger.Print(e.Error())
		}
		table.Add(vm.Name, vm.Id(), vm.State, started, vmx.Cpus, vmx.Memory, vmx.MacAddress, vm.IP, vmx.SoftPowerOff, vmx.CleanShutdown)
	}
	logger.Println(table)
	return nil
}

func runningVMs() ([]string, error) {
	return nil, nil
}
