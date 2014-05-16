package vbox

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

func DownloadTemplateVM(sourceURL, filename, vm string) (e error) {
	vmList, e := ListAllVMs()
	if e != nil {
		return e
	}

	for i := range vmList {
		if vmList[i].Name == vm {
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

func ListAllVMs() ([]*VM, error) {
	return listVMs(true)
}

func ListRunningVMs() ([]*VM, error) {
	return listVMs(false)
}

func listVMs(all bool) (vms []*VM, e error) {
	listType := "vms"
	if !all {
		listType = "runningvms"
	}

	var output []string
	if output, e = run("list", listType); e != nil {
		return nil, e
	}

	vms = make([]*VM, 0, len(output))

	for i := range output {
		if len(output[i]) == 0 {
			continue
		}

		parts := strings.Fields(output[i])
		if len(parts) != 2 {
			return nil, fmt.Errorf("failed to parse line: %s", output[i])
		}
		vm := &VM{Name: strings.Trim(parts[0], "\""), Uuid: parts[1]}
		if e = vm.load(); e != nil {
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

func GetVMProperties(vm string) (properties map[string]string, e error) {
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

func (vm *VM) load() (e error) {
	var output []string
	if output, e = run("showvminfo", "--machinereadable", vm.Name); e != nil {
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

	if vm.Cpus, e = strconv.Atoi(values["cpus"]); e != nil {
		return e
	}

	if vm.Memory, e = strconv.Atoi(values["memory"]); e != nil {
		return e
	}

	vm.Status = values["VMState"]

	for i := 0; i < 4; i++ {
		vm.BootOrder[i] = values[fmt.Sprintf("boot%d", i+1)]
	}

	for i := 0; i < 8; i++ {
		if ntype, found := values[fmt.Sprintf("nic%d", i+1)]; found {
			if ntype == "none" {
				continue
			}
			nic := &VNet{
				Id:    i + 1,
				NType: ntype,
			}
			nic.Mac = values[fmt.Sprintf("macaddress%d", nic.Id)]
			if ntype == "hostonly" {
				nic.Name = values[fmt.Sprintf("hostonlyadapter%d", nic.Id)]
			}
			vm.Nics = append(vm.Nics, nic)
		}
	}

	if vm.SFolders == nil {
		vm.SFolders = map[string]string{}
	}
	for i := 1; ; i++ {
		sfMap, found := values[fmt.Sprintf("SharedFolderNameMachineMapping%d", i)]

		if !found {
			break
		}

		vm.SFolders[sfMap] = values[fmt.Sprintf("SharedFolderPathMachineMapping%d", i)]
	}

	return nil
}

func CloneVM(name string, template string, snapshot string) (e error) {
	if _, e = run("clonevm", template, "--name", name, "--snapshot", snapshot, "--options", "link", "--register"); e != nil {
		return e
	}

	return e
}

type VM struct {
	Name      string
	Uuid      string
	Status    string
	BootOrder [4]string
	Memory    int
	Cpus      int
	Nics      []*VNet
	SFolders  map[string]string
}

type VNet struct {
	Id    int
	NType string
	Name  string
	Mac   string
}

func LoadVM(name string) (vm *VM, e error) {
	vm = &VM{Name: name}
	return vm, vm.load()
}

func StartVM(name string, withGui bool) (e error) {
	vmtype := "headless"
	if withGui {
		vmtype = "gui"
	}
	_, e = run("startvm", name, "--type", vmtype)
	return e
}

func SaveVM(name string) (e error) {
	_, e = run("controlvm", name, "savestate")
	return e
}

func StopVM(name string) (e error) {
	_, e = run("controlvm", name, "poweroff")
	return e
}

func ShutdownVM(name string) (e error) {
	_, e = run("controlvm", name, "acpipowerbutton")
	return e
}

func isVMRunning(name string) (bool, error) {
	vms, e := ListRunningVMs()
	if e != nil {
		return false, e
	}

	for i := range vms {
		if vms[i].Name == name {
			return true, nil
		}
	}
	return false, nil
}

func isVM(name string) (bool, error) {
	vms, e := ListAllVMs()
	if e != nil {
		return false, e
	}

	for i := range vms {
		if vms[i].Name == name {
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

func GetIP(vm string, iface int, timeout int) (ip string, e error) {
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

func DeleteVM(name string) (e error) {
	_, e = run("unregistervm", name, "--delete")
	return e
}

func (vm *VM) Save() (e error) {
	args := []string{vm.Name}
	args = append(args, "--memory", strconv.Itoa(vm.Memory))
	args = append(args, "--cpus", strconv.Itoa(vm.Cpus))

	for i := 0; i < 4; i++ {
		args = append(args, "--boot"+strconv.Itoa(i+1), vm.BootOrder[i])
	}

	for _, nic := range vm.Nics {
		args = append(args, "--nic"+strconv.Itoa(nic.Id), nic.NType)
		if nic.NType == "hostonly" {
			args = append(args, "--hostonlyadapter"+strconv.Itoa(nic.Id), nic.Name)
		}
	}

	_, e = run("modifyvm", args...)

	return e
}

func ShareFolder(name, tname, folder string) (e error) {
	if _, e = os.Stat(folder); os.IsNotExist(e) {
		return fmt.Errorf("folder %q does not exist!", folder)
	}

	_, e = run("sharedfolder", "add", name, "--name", tname, "--hostpath", folder, "--automount")
	return e
}

func UnshareFolder(name, tname string) (e error) {
	_, e = run("sharedfolder", "remove", name, "--name", tname)
	return e
}

func (vm *VM) String() string {
	s := ""
	s += fmt.Sprintf("VM %q\n", vm.Name)
	s += fmt.Sprintf("cpus:          %d\n", vm.Cpus)
	s += fmt.Sprintf("memory:        %d MB\n", vm.Memory)
	s += fmt.Sprintf("boot order:    %s\n", strings.Join(vm.BootOrder[:], ","))
	for _, nic := range vm.Nics {
		networkName := ""
		if nic.NType == "hostonly" {
			networkName = " [->" + nic.Name + "]"
		}
		s += fmt.Sprintf("nic [%d]:       %s %s%s\n", nic.Id, nic.NType, nic.Mac, networkName)
	}
	for k, v := range vm.SFolders {
		s += fmt.Sprintf("shared folder: %s [->%s]\n", v, k)
	}
	return s
}
