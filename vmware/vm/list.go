package main

import (
	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
	"log"
)

func init() {
	router.Register("list", &ListAction{})
}

type ListAction struct {
}

func (list *ListAction) Run() error {
	vms, e := vmware.All()
	if e != nil {
		return e
	}
	table := gocli.NewTable()
	leases, e := vmware.AllLeases()
	if e != nil {
		return e
	}
	table.Add("Name", "Status", "Mac", "Ip", "SoftPowerOff", "CleanShutdown")
	for _, vm := range vms {
		//client := gossh.New(ip, "root")
		//defer client.Close()
		//release, e := client.Execute("lsb_release -d")
		//if e != nil {
		//	return e
		//}
		//table.Add(vm.Name(), ip, strings.TrimSpace(release.Stdout()))
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
		table.Add(vm.Name(), status, mac, ip, vmx.SoftPowerOff, vmx.CleanShutdown)
	}
	log.Println(table)
	return nil
}
