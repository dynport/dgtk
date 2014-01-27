package main

import (
	"github.com/dynport/dgtk/cli"
	"log"
)

var vboxHost = &host{command: "VBoxManage"}

func main() {
	router := cli.NewRouter()

	router.Register("get/template", &getTemplateVM{}, "Get template VM used for cloning.")

	router.Register("vm/clone", &actCloneVM{}, "Clone a new VM from a template VM.")
	router.Register("vm/delete", &vmBase{Action: "delete"}, "Delete the VM with the given name.")

	router.Register("vm/start", &startVM{}, "Start the VM with the given name.")
	router.Register("vm/save", &vmBase{Action: "save"}, "Stop the VM with the given name (saving the current state).")
	router.Register("vm/stop", &vmBase{Action: "stop"}, "Stop the VM with the given name (unplug the VM).")

	router.Register("vm/props", &vmBase{Action: "props"}, "List properties of given VM.")

	router.Register("vm/ssh/into", &sshInto{}, "Connect to the VM using SSH.")

	router.Register("list", &actListVMs{}, "List available VMs.")

	if e := router.RunWithArgs(); e != nil {
		log.Fatal(e)
	}
}
