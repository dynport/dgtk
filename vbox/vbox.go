package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var VBoxManage = "VBoxManage"

func createVBoxError(output []string, e error) error {
	vboxErrorPrefix := "VBoxManage: error: "
	for i := range output {
		if strings.HasPrefix(output[i], vboxErrorPrefix) {
			return fmt.Errorf(strings.TrimPrefix(output[i], vboxErrorPrefix))
		}
	}
	return e
}

func downloadFile(baseUrl, filename string) (target string, e error) {
	target = filepath.Join(os.TempDir(), filename)

	if _, e = os.Stat(target); !os.IsNotExist(e) {
		return "", fmt.Errorf("target %q already exists", target)
	}

	out, e := os.Create(target)
	if e != nil {
		log.Printf("failed to create %q", target)
		return "", e
	}
	defer out.Close()

	url := baseUrl + "/" + filename

	resp, e := http.Get(url)
	if e != nil || resp.StatusCode != 200 {
		if resp == nil {
			return "", e
		}
		return "", fmt.Errorf("request for %q failed: %s", url, resp.Status)
	}
	defer resp.Body.Close()

	log.Print("downloading template ...")
	_, e = io.Copy(out, resp.Body)
	log.Print(" ... done")

	return target, nil
}

func downloadTemplateVM(sourceURL, filename, vm string) (e error) {
	vmList, e := listAllVMs()
	if e != nil {
		return e
	}

	for i := range vmList {
		if vmList[i].name == vm {
			return fmt.Errorf("vm %q already exists!", vm)
		}
	}

	target, e := downloadFile(sourceURL, filename)
	if e != nil {
		return e
	}
	defer os.Remove(target)

	log.Print("importing vm ...")
	_, e = run("import", target, "--vsys", "0", "--vmname", vm)
	log.Print(" ... done")
	return e
}

func run(action string, args ...string) (output []string, e error) {
	args = append([]string{action}, args...)
	cmd := exec.Command(VBoxManage, args...)

	out, e := cmd.CombinedOutput()
	if out != nil {
		output = strings.Split(string(out), "\n")
		if e != nil {
			return output, createVBoxError(output, e)
		}
	}
	return output, e
}

func listAllVMs() ([]*vbox, error) {
	return listVMs(true)
}

func listRunningVMs() ([]*vbox, error) {
	return listVMs(false)
}

func listVMs(all bool) (vms []*vbox, e error) {
	listType := "vms"
	if !all {
		listType = "runningvms"
	}

	var output []string
	if output, e = run("list", listType); e != nil {
		return nil, e
	}

	vms = make([]*vbox, 0, len(output))

	for i := range output {
		if len(output[i]) == 0 {
			continue
		}

		parts := strings.Fields(output[i])
		if len(parts) != 2 {
			return nil, fmt.Errorf("failed to parse line: %s", output[i])
		}
		vm := &vbox{name: strings.Trim(parts[0], "\""), uuid: parts[1]}
		if e = vmInfos(vm); e != nil {
			return nil, e
		}
		vms = append(vms, vm)
	}
	return vms, nil
}

func parsePropertyLine(line string) (name string, value string) {
	parts := strings.Split(line, ", ")

	name = strings.TrimPrefix(parts[0], "Name: ")
	value = strings.TrimPrefix(parts[1], "value: ")

	return name, value
}

func getVMProperties(vm string) (properties map[string]string, e error) {
	var output []string
	if output, e = run("guestproperty", "enumerate", vm); e != nil {
		return nil, e
	}

	properties = map[string]string{}
	for _, line := range output {
		if len(line) == 0 {
			continue
		}

		name, value := parsePropertyLine(line)
		properties[name] = value
	}

	return properties, nil
}

func vmInfos(vm *vbox) (e error) {
	var output []string
	if output, e = run("showvminfo", "--machinereadable", vm.name); e != nil {
		return e
	}

	values := map[string]string{}

	for _, line := range output {
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, "=")
		values[parts[0]] = strings.Trim(parts[1], "\"")
	}

	if vm.cpus, e = strconv.Atoi(values["cpus"]); e != nil {
		return e
	}

	if vm.memory, e = strconv.Atoi(values["memory"]); e != nil {
		return e
	}

	vm.status = values["VMState"]

	for i := 0; i < 4; i++ {
		vm.bootOrder[i] = values[fmt.Sprintf("boot%d", i+1)]
	}

	for i := 0; i < 8; i++ {
		if ntype, found := values[fmt.Sprintf("nic%d", i+1)]; found {
			if ntype == "none" {
				continue
			}
			nic := &vnet{
				id:    i + 1,
				ntype: ntype,
			}
			nic.mac = values[fmt.Sprintf("macaddress%d", nic.id)]
			if ntype == "hostonly" {
				nic.name = values[fmt.Sprintf("hostonlyadapter%d", nic.id)]
			}
			vm.nics = append(vm.nics, nic)
		}
	}

	return nil
}

func cloneVM(name string, template string, snapshot string) (e error) {
	if _, e = run("clonevm", template, "--name", name, "--snapshot", snapshot, "--options", "link", "--register"); e != nil {
		return e
	}

	return e
}

type vbox struct {
	name      string
	uuid      string
	status    string
	bootOrder [4]string
	memory    int
	cpus      int
	nics      []*vnet
}

type vnet struct {
	id    int
	ntype string
	name  string
	mac   string
}

func startVM(name string, withGui bool) (e error) {
	vmtype := "headless"
	if withGui {
		vmtype = "gui"
	}
	_, e = run("startvm", name, "--type", vmtype)
	return e
}

func saveVM(name string) (e error) {
	_, e = run("controlvm", name, "savestate")
	return e
}

func stopVM(name string) (e error) {
	_, e = run("controlvm", name, "poweroff")
	return e
}

func shutdownVM(name string) (e error) {
	_, e = run("controlvm", name, "acpipowerbutton")
	return e
}

func isVMRunning(name string) (bool, error) {
	vms, e := listRunningVMs()
	if e != nil {
		return false, e
	}

	for i := range vms {
		if vms[i].name == name {
			return true, nil
		}
	}
	return false, nil
}

func isVM(name string) (bool, error) {
	vms, e := listAllVMs()
	if e != nil {
		return false, e
	}

	for i := range vms {
		if vms[i].name == name {
			return true, nil
		}
	}
	return false, nil
}

func getProperty(vm string, property string) (value string, e error) {
	result, e := run("guestproperty", "get", vm, property)
	if e != nil {
		return "", e
	}

	if result[0] == "No value set!" {
		return "", nil
	}

	if strings.HasPrefix(result[0], "Value: ") {
		return strings.TrimPrefix(result[0], "Value: "), nil
	}

	return "", fmt.Errorf("failed to retrieve property")
}

func waitForProperty(vm string, property string, timeout int) (value string, e error) {
	result, e := run("guestproperty", "wait", vm, property, "--timeout", strconv.Itoa(1000*timeout))
	if e != nil {
		return "", e
	} else if strings.HasPrefix(result[0], "VBoxManage: error: Time out or interruption while waiting for a notification.") {
		return "", fmt.Errorf("Machine not reachable.")
	}

	_, ip := parsePropertyLine(result[0])
	return ip, nil
}

func getIP(vm string, iface int, timeout int) (ip string, e error) {
	var valid, running bool
	if valid, e = isVM(vm); e != nil {
		return "", e
	} else if !valid {
		return "", fmt.Errorf("VM %q does not exist", vm)
	}

	if running, e = isVMRunning(vm); e != nil {
		return "", e
	} else if !running {
		return "", fmt.Errorf("VM %q not running", vm)
	}

	property := fmt.Sprintf("/VirtualBox/GuestInfo/Net/%d/V4/IP", iface)

	ip, e = getProperty(vm, property)
	if e != nil {
		return "", e
	} else if ip == "" {
		ip, e = waitForProperty(vm, property, timeout)
		if e != nil {
			return "", e
		}
	}
	return ip, nil
}

func deleteVM(name string) (e error) {
	_, e = run("unregistervm", name, "--delete")
	return e
}

func configureVM(vm *vbox) (e error) {
	args := []string{vm.name}
	args = append(args, "--memory", strconv.Itoa(vm.memory))
	args = append(args, "--cpus", strconv.Itoa(vm.cpus))

	for i := 0; i < 4; i++ {
		args = append(args, "--boot"+strconv.Itoa(i+1), vm.bootOrder[i])
	}

	for _, nic := range vm.nics {
		args = append(args, "--nic"+strconv.Itoa(nic.id), nic.ntype)
		if nic.ntype == "hostonly" {
			args = append(args, "--hostonlyadapter"+strconv.Itoa(nic.id), nic.name)
		}
	}

	_, e = run("modifyvm", args...)

	return e
}

func shareFolder(name, tname, folder string) (e error) {
	_, e = run("sharefolder", "add", name, "--name", tname, "--hostpath", folder, "--automount")
	return e
}

func unshareFolder(name, tname string) (e error) {
	_, e = run("sharefolder", "remove", name, "--name", tname)
	return e
}
