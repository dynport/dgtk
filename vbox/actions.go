package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type actDownloadTemplateVM struct {
	SourceURL  string `cli:"type=opt short=s long=source default='http://192.168.1.10/vbox' desc='location of the template to download'"`
	Template   string `cli:"type=arg required=true desc='The name of the template to load (try ubuntu_precise_template.ova)'"`
	Identifier string `cli:"type=arg required=true desc='Identifier of the template VM'"`
}

func (action *actDownloadTemplateVM) Run() error {
	return downloadTemplateVM(action.SourceURL, action.Template, action.Identifier)
}

type actCloneVM struct {
	Template string `cli:"type=opt short=t long=template default='template' desc='The VM to use for cloning.'"`
	Snapshot string `cli:"type=opt short=s long=snapshot default='base' desc='The template VM\'s snapshot to use for cloning.'"`
	Name     string `cli:"type=arg required=true desc='Name of the new VM'"`
}

func (action *actCloneVM) Run() error {
	return cloneVM(action.Name, action.Template, action.Snapshot)
}

type actListVMs struct {
	Running bool `cli:"type=opt short=r long=running desc='Show running VMs only.'"`
}

func (action *actListVMs) Run() (e error) {
	var vms []*vbox
	if action.Running {
		vms, e = listRunningVMs()
	} else {
		vms, e = listAllVMs()
	}
	if e != nil {
		return e
	}

	for _, vm := range vms {
		r := ""
		if vm.status == "running" {
			r = "*"
		}
		log.Printf("%s%s", vm.name, r)
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
		return stopVM(action.Name)
	case "shutdown":
		return shutdownVM(action.Name)
	case "save":
		return saveVM(action.Name)
	case "delete":
		return deleteVM(action.Name)
	case "info":
		return showVMInfo(action.Name)
	}
	return nil
}

func listVMProps(vm string) (e error) {
	var props map[string]string
	if props, e = getVMProperties(vm); e != nil {
		return e
	}

	for k, v := range props {
		if !strings.HasPrefix(k, "/VirtualBox/GuestInfo") {
			continue
		}
		log.Printf("%s%-*s%q", k, 50-len(k), "", v)
	}
	return nil
}

func showVMInfo(name string) (e error) {
	vm := &vbox{name: name}
	if e = vmInfos(vm); e != nil {
		return e
	}
	log.Printf("VM %q", name)
	log.Printf("cpus:       %d", vm.cpus)
	log.Printf("memory:     %d kB", vm.memory)
	log.Printf("boot order: %s", strings.Join(vm.bootOrder[:], ","))
	return nil
}

type actStartVM struct {
	vmBase
	WithGUI bool `cli:"type=opt short=g long=gui desc='Show the GUI of the virtual machine?'"`
}

func (action *actStartVM) Run() (e error) {
	return startVM(action.Name, action.WithGUI)
}

type sshInto struct {
	vmBase

	User    string `cli:"type=opt short=l long=login default=root desc='User used to log in to the VM.'"`
	IFace   int    `cli:"type=opt short=i long=interface default=0 desc='Number of the nic to connect to.'"`
	Timeout int    `cli:"type=opt short=t long=timeout default=15 desc='Time to wait for machine to boot.'"`
}

func (action *sshInto) Run() error {
	ip, e := getIP(action.Name, action.IFace, action.Timeout)
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

type actConfigureVM struct {
	vmBase

	CPUs      int    `cli:"opt -c --cpus default=-1 desc='Change the number of CPUs of the VM.'"`
	Memory    int    `cli:"opt -m --memory default=-1 desc='Change the amount of memory the VM has.'"`
	BootOrder string `cli:"opt -b --boot-order desc='Comma separated list of devices used for boot from floppy, dvd, disk or net'"`
}

func (action *actConfigureVM) Run() (e error) {
	vm := &vbox{name: action.Name}
	if e = vmInfos(vm); e != nil {
		return e
	}

	if action.CPUs != -1 {
		vm.cpus = action.CPUs
	}

	if action.Memory != -1 {
		vm.memory = action.Memory
	}

	if action.BootOrder != "" {
		devices := strings.Split(action.BootOrder, ",")
		for i := 0; i < 4; i++ {
			if i < len(devices) {
				vm.bootOrder[i] = strings.TrimSpace(devices[i])
			} else {
				vm.bootOrder[i] = "none"
			}
		}
	}

	return configureVM(vm)
}
