package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dynport/dgtk/vbox"
)

type actImportTemplateVM struct {
	FileToImport string `cli:"arg required desc='Filename of the template VM to import'"`
	Identifier   string `cli:"arg required desc='Identifier of the template VM'"`
}

func (action *actImportTemplateVM) Run() error {
	return vbox.ImportTemplateVM(action.FileToImport, action.Identifier)
}

type actDownloadTemplateVM struct {
	SourceURL  string `cli:"type=opt short=s long=source default='http://192.168.1.10/vbox' desc='location of the template to download'"`
	Template   string `cli:"type=arg required=true desc='The name of the template to load (try ubuntu_precise_template.ova)'"`
	Identifier string `cli:"type=arg required=true desc='Identifier of the template VM'"`
}

func (action *actDownloadTemplateVM) Run() error {
	return vbox.DownloadTemplateVM(action.SourceURL, action.Template, action.Identifier)
}

type actCloneVM struct {
	Snapshot string `cli:"type=opt short=s long=snapshot default='base' desc='The template VMs snapshot to use for cloning.'"`
	Template string `cli:"type=arg required=true desc='The VM to use for cloning.'"`
	Name     string `cli:"type=arg required=true desc='Name of the new VM'"`
}

func (action *actCloneVM) Run() error {
	return vbox.CloneVM(action.Name, action.Template, action.Snapshot)
}

type actListVMs struct {
	Running bool `cli:"type=opt short=r long=running desc='Show running VMs only.'"`
}

func (action *actListVMs) Run() (e error) {
	var vms []*vbox.VM
	if action.Running {
		vms, e = vbox.ListRunningVMs()
	} else {
		vms, e = vbox.ListAllVMs()
	}
	if e != nil {
		return e
	}

	for _, vm := range vms {
		r := ""
		if vm.Status == "running" {
			r = "*"
		}
		fmt.Printf("%s%s\n", vm.Name, r)
	}
	return nil
}

type vmBase struct {
	Name   string `cli:"type=arg required=true desc='Name of the VM'"`
	Action string
}

func (action *vmBase) Run() (e error) {
	switch action.Action {
	case "props":
		return listVMProps(action.Name)
	case "stop":
		return vbox.StopVM(action.Name)
	case "shutdown":
		return vbox.ShutdownVM(action.Name)
	case "save":
		return vbox.SaveVM(action.Name)
	case "delete":
		return vbox.DeleteVM(action.Name)
	case "info":
		return showVMInfo(action.Name)
	}
	return nil
}

func listVMProps(vm string) (e error) {
	var props map[string]string
	if props, e = vbox.GetVMProperties(vm); e != nil {
		return e
	}

	for k, v := range props {
		if !strings.HasPrefix(k, "/VirtualBox/GuestInfo") {
			continue
		}
		fmt.Printf("%s%-*s%q\n", k, 50-len(k), "", v)
	}
	return nil
}

func showVMInfo(name string) (e error) {
	vm, e := vbox.LoadVM(name)
	if e != nil {
		return e
	}

	log.Printf("%s", vm)

	return nil
}

type actStartVM struct {
	vmBase
	WithGUI bool `cli:"type=opt short=g long=gui desc='Show the GUI of the virtual machine?'"`
}

func (action *actStartVM) Run() (e error) {
	return vbox.StartVM(action.Name, action.WithGUI)
}

type sshInto struct {
	vmBase

	User    string `cli:"type=opt short=l long=login default=ubuntu desc='User used to log in to the VM.'"`
	IFace   int    `cli:"type=opt short=i long=interface default=0 desc='Number of the nic to connect to.'"`
	Timeout int    `cli:"type=opt short=t long=timeout default=15 desc='Time to wait for machine to boot.'"`
}

func (action *sshInto) Run() error {
	ip, e := vbox.GetIP(action.Name, action.IFace, action.Timeout)
	if e != nil {
		return e
	}

	log.Printf("connecting to machine %q using ip %q", action.Name, ip)

	cmd := exec.Command("ssh", "-l", action.User, ip)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type actShareFolder struct {
	vmBase

	LocalPath  string `cli:"arg required desc='Absolute path of the folder shared with the guest.'"`
	RemoteName string `cli:"arg required desc='Remote name of the path (mounted in /media).'"`
}

func (action *actShareFolder) Run() (e error) {
	path := action.LocalPath
	if !filepath.IsAbs(path) {
		if path, e = filepath.Abs(path); e != nil {
			return e
		}
	}

	return vbox.ShareFolder(action.Name, action.RemoteName, path)
}

type actUnshareFolder struct {
	vmBase

	RemoteName string `cli:"arg required desc='Remote name of the path (mounted in /media).'"`
}

func (action *actUnshareFolder) Run() error {
	return vbox.UnshareFolder(action.Name, action.RemoteName)
}

type actConfigMemVM struct {
	vmBase

	Memory int `cli:"arg required desc='The amount of memory the VM has in MB.'"`
}

func (action *actConfigMemVM) Run() (e error) {
	vm, e := vbox.LoadVM(action.Name)
	if e != nil {
		return e
	}

	vm.Memory = action.Memory
	return vm.Save()
}

type actConfigCPUsVM struct {
	vmBase

	CPUs int `cli:"arg required desc='The amount of CPUs the VM has.'"`
}

func (action *actConfigCPUsVM) Run() (e error) {
	vm, e := vbox.LoadVM(action.Name)
	if e != nil {
		return e
	}

	vm.Cpus = action.CPUs

	return vm.Save()
}

type actConfigBootOrderVM struct {
	vmBase

	Devices []string `cli:"arg required desc='Ordered list of boot devices (at most 4 of floppy, disk, dvd, and net)'"`
}

func (action *actConfigBootOrderVM) Run() (e error) {
	vm, e := vbox.LoadVM(action.Name)
	if e != nil {
		return e
	}

	for i := 0; i < 4; i++ {
		if i < len(action.Devices) {
			vm.BootOrder[i] = action.Devices[i]
		} else {
			vm.BootOrder[i] = "none"
		}
	}

	return vm.Save()
}

type actConfigNetworkIFacesVM struct {
	vmBase

	Id      int    `cli:"arg required desc='Network interface to configure.'"`
	NType   string `cli:"arg required desc='Type of NIC (one of none, bridged, nat, and hostonly).'"`
	Network string `cli:"arg desc='Network name (required for hostonly).'"`
}

func (action *actConfigNetworkIFacesVM) Run() (e error) {
	vm, e := vbox.LoadVM(action.Name)
	if e != nil {
		return e
	}

	if action.Id >= 8 {
		return fmt.Errorf("Only 8 nics supported!")
	}

	var nic *vbox.VNet
	for i := range vm.Nics {
		if vm.Nics[i].Id == action.Id {
			nic = vm.Nics[i]
		}
	}

	if nic == nil {
		nic = &vbox.VNet{Id: action.Id}
		vm.Nics = append(vm.Nics, nic)
	}

	nic.NType = action.NType
	if action.NType == "hostonly" {
		if action.Network == "" {
			return fmt.Errorf("No name for the hostonly network specified!")
		}
		nic.Name = action.Network
	}

	return vm.Save()
}
