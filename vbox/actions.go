package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type getTemplateVM struct {
	SourceURL  string `cli:"type=opt short=s long=source default='http://192.168.1.10/vbox' desc='location of the template to download'"`
	Template   string `cli:"type=arg required=true desc='The name of the template to load (try ubuntu_precise_template.ova)'"`
	Identifier string `cli:"type=arg required=true desc='Identifier of the template VM'"`
}

func (action *getTemplateVM) Run() error {
	return vboxHost.getTemplateVM(action.SourceURL, action.Template, action.Identifier)
}

type actCloneVM struct {
	Template string `cli:"type=opt short=t long=template default='template' desc='The VM to use for cloning.'"`
	Snapshot string `cli:"type=opt short=s long=snapshot default='base' desc='The template VM\'s snapshot to use for cloning.'"`
	Name     string `cli:"type=arg required=true desc='Name of the new VM'"`
}

func (action *actCloneVM) Run() error {
	return vboxHost.cloneVM(action.Name, action.Template, action.Snapshot)
}

func listVMs() (e error) {
	var vms []*vm
	if vms, e = vboxHost.listAllVMs(); e != nil {
		return e
	}
	for _, vm := range vms {
		log.Printf("%s", vm.name)
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
		return vboxHost.stopVM(action.Name)
	case "save":
		return vboxHost.saveVM(action.Name)
	case "delete":
		return vboxHost.deleteVM(action.Name)
	}
	return nil
}

func listVMProps(vm string) (e error) {
	var props map[string]string
	if props, e = vboxHost.getVMProperties(vm); e != nil {
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

type startVM struct {
	vmBase
	WithGUI bool `cli:"type=opt short=g long=gui desc='Show the GUI of the virtual machine?'"`
}

func (action *startVM) Run() (e error) {
	return vboxHost.startVM(action.Name, action.WithGUI)
}

type sshInto struct {
	vmBase

	User    string `cli:"type=opt short=u long=user default=root desc='User used to log in to the VM.'"`
	IFace   int    `cli:"type=opt short=i long=interface default=0 desc='Number of the nic to connect to.'"`
	Timeout int    `cli:"type=opt short=t long=timeout default=15 desc='Time to wait for machine to boot.'"`
}

func (action *sshInto) Run() error {
	ip, e := vboxHost.getIP(action.Name, action.IFace, action.Timeout)
	if e != nil {
		return e
	}

	cmd := exec.Command("ssh", "-l", action.User, ip)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
