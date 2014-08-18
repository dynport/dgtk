package vmware

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var debug = os.Getenv("DEBUG") == "true"

func Start(path string, gui bool) error {
	guiFlag := "nogui"
	if gui {
		guiFlag = "gui"
	}
	_, e := vmrun("start", path, guiFlag)
	if e != nil {
		return e
	}
	return nil
}

func DeleteSnapshot(path string, name string) error {
	out, e := vmrun("deleteSnapshot", path, name)
	logger.Println(out)
	return e
}

func RestoreSnapshot(path string, name string) error {
	out, e := vmrun("revertToSnapshot", path, name)
	logger.Println(out)
	return e
}

func TakeSnapshot(path string, name string) error {
	out, e := vmrun("snapshot", path, name)
	logger.Println(out)
	return e
}

func ListSnapshots(path string) (snapshots []*Snapshot, e error) {
	out, e := vmrun("listSnapshots", path)
	if e != nil {
		return nil, e
	}
	lines := strings.Split(out, "\n")
	if len(lines) > 1 {
		for _, line := range lines[1:] {
			snapshots = append(snapshots, &Snapshot{Name: line})
		}
	}
	return snapshots, nil
}

func Stop(path string) error {
	_, e := vmrun("stop", path)
	if e != nil {
		return e
	}
	return nil
}

type CloneOptions struct {
	Linked    bool
	Snapshot  string
	CloneName string
}

func DeleteVM(path_ string) error {
	_, e := vmrun("deleteVM", path_)
	if e != nil {
		return e
	}
	dir := path.Dir(path_)
	logger.Printf("removing dir %q", dir)
	return os.RemoveAll(dir)
}

func randId() string {
	id := "vm-"
	for i := 0; i < 8; i++ {
		id += fmt.Sprintf("%x", rand.Int31n(16))
	}
	return id
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func Create(vm *Vm, snapshot string) (*Vm, error) {
	if vm == nil {
		return nil, fmt.Errorf("provided VM is nil")
	}
	id := randId()
	dst := CloudVMLocation + "/" + id + "/" + id + ".vmx"
	e := os.MkdirAll(path.Dir(dst), 0755)
	if e != nil {
		return nil, e
	}
	logger.Printf("cloning %q to %q", vm.Id(), dst)
	started := time.Now()
	vm, e = Clone(vm.Path, dst, &CloneOptions{Snapshot: snapshot, CloneName: id})
	if e != nil {
		return nil, e
	}
	logger.Printf("cloned in %.3f", time.Since(started).Seconds())
	return vm, nil
}

func Clone(src string, dst string, opts *CloneOptions) (*Vm, error) {
	typ := "full"
	if opts.Linked {
		typ = "linked"
	}
	args := []string{src, dst, typ}
	if opts.Snapshot != "" {
		args = append(args, "-snapshot="+opts.Snapshot)
	}
	if opts.CloneName != "" {
		args = append(args, "-cloneName="+opts.CloneName)
	}
	_, e := vmrun("clone", args...)
	return &Vm{Path: dst}, e
}

func vmrun(vmrunCmd string, params ...string) (string, error) {
	args := append([]string{"-T", "fusion", vmrunCmd}, params...)
	if debug {
		logger.Printf("DEBUG: %v", args)
	}
	cmd := exec.Command("/Applications/VMware Fusion.app/Contents/Library/vmrun", args...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return "", fmt.Errorf(e.Error() + " " + string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func GetGuestIPAddress(vmx string) (string, error) {
	return vmrun("getGuestIPAddress", vmx)
}

var (
	DefaultVMLocation = os.Getenv("HOME") + "/Documents/Virtual Machines.localized"
	Root              = os.Getenv("HOME") + "/.vmware"
	CloudVMLocation   = Root + "/vms"
	TemplatesPath     = Root + "/templates"
)

func AllTemplates() (vms Vms, e error) {
	return glob(TemplatesPath)
}

func AllWithTemplates() (vms Vms, e error) {
	return FindVms([]string{DefaultVMLocation, CloudVMLocation, TemplatesPath})
}

func All() (vms Vms, e error) {
	return FindVms([]string{DefaultVMLocation, CloudVMLocation})
}

func FindVms(locations []string) (vms Vms, e error) {
	raw, e := List()
	if e != nil {
		return nil, e
	}

	for _, l := range locations {
		if tmp, e := glob(l); e == nil {
			raw = append(raw, tmp...)
		}
	}
	m := map[string]bool{}
	for _, a := range raw {
		if _, ok := m[a.Path]; !ok {
			vms = append(vms, a)
			m[a.Path] = true

		}
	}
	return vms, nil
}

func glob(path string) (vms Vms, e error) {
	files, e := filepath.Glob(path + "/*/*.vmx")
	if e != nil {
		return nil, e
	}
	for _, f := range files {
		vms = append(vms, &Vm{Path: f})
	}
	return vms, nil
}

func List() (vms Vms, e error) {
	out, e := vmrun("list")
	if e != nil {
		return nil, e
	}
	dbg.Printf("%s", out)
	lines := strings.Split(out, "\n")
	if len(lines) > 1 {
		for _, line := range lines[1:] {
			vms = append(vms, &Vm{Path: line})
		}
	}
	return vms, nil
}
