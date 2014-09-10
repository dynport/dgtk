package main

import (
	"sort"

	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
)

type ListAction struct {
}

func (list *ListAction) Run() error {
	vms, e := vmware.All()
	if e != nil {
		return e
	}
	sort.Sort(vms)
	table := gocli.NewTable()
	leases, e := vmware.AllLeases()
	if e != nil {
		return e
	}
	tags, e := vmware.LoadTags()
	if e != nil {
		return e
	}

	tagsMap := map[string]string{}
	for _, t := range tags {
		tagsMap[t.Id()] = t.Value
	}
	table.Add("Id", "Name", "Status", "Started", "Cpus", "Memory", "Mac", "Ip", "SoftPowerOff", "CleanShutdown")
	for _, vm := range vms {
		vmx, e := vm.Vmx()
		if e != nil {
			return e
		}
		mac := vmx.MacAddress
		lease := leases.Lookup(mac)
		ip := ""
		if lease != nil {
			ip = lease.Ip
		}
		status := "STOPPED"
		if vm.Running() {
			status = "RUNNING"
		}
		started := ""
		if s, e := vm.StartedAt(); e == nil {
			started = s.Format("2006-01-02T15:04:05")
		} else {
			logger.Print(e.Error())
		}
		name := tagsMap[vm.Id()+":Name"]
		table.Add(name, vm.Id(), status, started, vmx.Cpus, vmx.Memory, mac, ip, vmx.SoftPowerOff, vmx.CleanShutdown)
	}
	logger.Println(table)
	return nil
}
