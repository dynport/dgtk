package main

import (
	"log"
	"os"
	"sort"

	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
)

type ListAction struct {
}

var logger = log.New(os.Stderr, "", 0)

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
	tags := &vmware.Tags{}
	e = tags.Load()
	if e != nil {
		return e
	}
	tagsMap := map[string]string{}
	for _, t := range tags.Tags() {
		tagsMap[t.Id()] = t.Value
	}
	logger.Printf("%#v", tags.Len())
	table.Add("Id", "Name", "Status", "Started", "Mac", "Ip", "SoftPowerOff", "CleanShutdown")
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
		name := tagsMap[vm.Id()+":Name"]
		table.Add(name, vm.Id(), status, started, mac, ip, vmx.SoftPowerOff, vmx.CleanShutdown)
	}
	log.Println(table)
	return nil
}
