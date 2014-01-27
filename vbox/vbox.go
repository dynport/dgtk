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

type host struct {
	command string
}

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

func (h *host) getTemplateVM(sourceURL, filename, vm string) (e error) {
	vmList, e := h.listAllVMs()
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
	_, e = h.run("import", target, "--vsys", "0", "--vmname", vm)
	log.Print(" ... done")
	return e
}

func (h *host) run(action string, args ...string) (output []string, e error) {
	args = append([]string{action}, args...)
	cmd := exec.Command(h.command, args...)

	out, e := cmd.CombinedOutput()
	if out != nil {
		output = strings.Split(string(out), "\n")
		if e != nil {
			return output, createVBoxError(output, e)
		}
	}
	return output, e
}

func (h *host) listAllVMs() ([]*vm, error) {
	return h.listVMs(true)
}

func (h *host) listRunningVMs() ([]*vm, error) {
	return h.listVMs(false)
}

func (h *host) listVMs(all bool) (vms []*vm, e error) {
	listType := "vms"
	if !all {
		listType = "runningvms"
	}

	var output []string
	if output, e = h.run("list", listType); e != nil {
		return nil, e
	}

	vms = make([]*vm, 0, len(output))

	for i := range output {
		if len(output[i]) == 0 {
			continue
		}

		parts := strings.Fields(output[i])
		if len(parts) != 2 {
			return nil, fmt.Errorf("failed to parse line: %s", output[i])
		}
		vms = append(vms, &vm{name: strings.Trim(parts[0], "\""), uuid: parts[1]})
	}
	return vms, nil
}

func parsePropertyLine(line string) (name string, value string) {
	parts := strings.Split(line, ", ")

	name = strings.TrimPrefix(parts[0], "Name: ")
	value = strings.TrimPrefix(parts[1], "value: ")

	return name, value
}

func (h *host) getVMProperties(vm string) (properties map[string]string, e error) {
	var output []string
	if output, e = h.run("guestproperty", "enumerate", vm); e != nil {
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

func (h *host) cloneVM(name string, template string, snapshot string) (e error) {
	if _, e = h.run("clonevm", template, "--name", name, "--snapshot", snapshot, "--options", "link", "--register"); e != nil {
		return e
	}

	return h.startVM(name, false)
}

type vm struct {
	name string
	uuid string
}

func (h *host) startVM(name string, withGui bool) (e error) {
	vmtype := "headless"
	if withGui {
		vmtype = "gui"
	}
	_, e = h.run("startvm", name, "--type", vmtype)
	return e
}

func (h *host) saveVM(name string) (e error) {
	_, e = h.run("controlvm", name, "savestate")
	return e
}

func (h *host) stopVM(name string) (e error) {
	_, e = h.run("controlvm", name, "poweroff")
	return e
}

func (h *host) isVMRunning(name string) (bool, error) {
	vms, e := h.listRunningVMs()
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

func (h *host) isVM(name string) (bool, error) {
	vms, e := h.listAllVMs()
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

func (h *host) getProperty(vm string, property string) (value string, e error) {
	result, e := h.run("guestproperty", "get", vm, property)
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

func (h *host) waitForProperty(vm string, property string, timeout int) (value string, e error) {
	result, e := h.run("guestproperty", "wait", vm, property, "--timeout", strconv.Itoa(1000*timeout))
	if e != nil {
		return "", e
	} else if strings.HasPrefix(result[0], "Time out or interruption while waiting.") {
		return "", fmt.Errorf("Machine not reachable.")
	}

	_, ip := parsePropertyLine(result[0])
	return ip, nil
}

func (h *host) getIP(vm string, iface int, timeout int) (ip string, e error) {
	var valid, running bool
	if valid, e = h.isVM(vm); e != nil {
		return "", e
	} else if !valid {
		return "", fmt.Errorf("VM %q does not exist", vm)
	}

	if running, e = h.isVMRunning(vm); e != nil {
		return "", e
	} else if !running {
		return "", fmt.Errorf("VM %q not running", vm)
	}

	property := fmt.Sprintf("/VirtualBox/GuestInfo/Net/%d/V4/IP", iface)

	ip, e = h.getProperty(vm, property)
	if e != nil {
		return "", e
	} else if ip == "" {
		ip, e = h.waitForProperty(vm, property, timeout)
		if e != nil {
			return "", e
		}
	}
	return ip, nil
}

func (h *host) deleteVM(name string) (e error) {
	_, e = h.run("unregistervm", name, "--delete")
	return e
}
