package main

import (
	"log"

	"github.com/dynport/dgtk/cli"
)

func main() {
	router := cli.NewRouter()

	router.Register("get/template", &actDownloadTemplateVM{}, "Get template VM used for cloning.")

	router.Register("list", &actListVMs{}, "List available VMs.")

	router.Register("vm/info", &vmBase{Action: "info"}, "Show information on the given VM.")

	router.Register("vm/config/mem", &actConfigMemVM{}, "Configure the VM to have given amount of memory.")
	router.Register("vm/config/cpu", &actConfigCPUsVM{}, "Configure the VM to have given amount of CPUs.")
	router.Register("vm/config/boot", &actConfigBootOrderVM{}, "Configure the boot order of the VM.")
	router.Register("vm/config/nic", &actConfigNetworkIFacesVM{}, "Configure the according NIC of the VM.")

	router.Register("vm/share", &actShareFolder{}, "Share a folder with the given VM.")
	router.Register("vm/unshare", &actUnshareFolder{}, "Unshare a folder with the given VM.")

	router.Register("clone", &actCloneVM{}, "Clone a new VM from a template VM.")
	router.Register("delete", &vmBase{Action: "delete"}, "Delete the VM with the given name.")

	router.Register("start", &actStartVM{}, "Start the VM with the given name.")
	router.Register("save", &vmBase{Action: "save"}, "Stop the VM with the given name (saving the current state).")
	router.Register("stop", &vmBase{Action: "stop"}, "Stop the VM with the given name (unplug the VM).")
	router.Register("shutdown", &vmBase{Action: "shutdown"}, "Send the VM the ACPI shutdown signal.")

	router.Register("vm/props", &vmBase{Action: "props"}, "List properties of given VM.")

	router.Register("ssh/into", &sshInto{}, "Connect to the VM using SSH.")

	if e := router.RunWithArgs(); e != nil {
		log.Fatal(e)
	}
}
