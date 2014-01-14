package main

import (
	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
	"log"
	"sort"
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
	table.Add("Name", "Status", "Started", "Mac", "Ip", "SoftPowerOff", "CleanShutdown")
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
			log.Print(e.Error())
		}
		table.Add(vm.Name(), status, started, mac, ip, vmx.SoftPowerOff, vmx.CleanShutdown)
	}
	log.Println(table)
	return nil
}
