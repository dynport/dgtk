package main

import (
	"log"

	"github.com/dynport/dgtk/cli"
)

func main() {
	router := cli.NewRouter()

	router.Register("get/template", &actDownloadTemplateVM{}, "Get template VM used for cloning.")

	router.Register("vm/list", &actListVMs{}, "List available VMs.")
	router.Register("vm/info", &vmBase{Action: "info"}, "Show information on the given VM.")
	router.Register("vm/configure ", &actConfigureVM{}, "Configure the given VM.")

	router.Register("vm/clone", &actCloneVM{}, "Clone a new VM from a template VM.")
	router.Register("vm/delete", &vmBase{Action: "delete"}, "Delete the VM with the given name.")

	router.Register("vm/start", &actStartVM{}, "Start the VM with the given name.")
	router.Register("vm/save", &vmBase{Action: "save"}, "Stop the VM with the given name (saving the current state).")
	router.Register("vm/stop", &vmBase{Action: "stop"}, "Stop the VM with the given name (unplug the VM).")
	router.Register("vm/shutdown", &vmBase{Action: "shutdown"}, "Send the VM the ACPI shutdown signal.")

	router.Register("vm/props", &vmBase{Action: "props"}, "List properties of given VM.")

	router.Register("vm/ssh/into", &sshInto{}, "Connect to the VM using SSH.")

	if e := router.RunWithArgs(); e != nil {
		log.Fatal(e)
	}
}
