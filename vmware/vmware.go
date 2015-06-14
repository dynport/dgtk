package vmware

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
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

func Running() ([]string, error) {
	b, err := vmrun("list")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	o := []string{}
	for i, l := range lines {
		if i > 0 {
			o = append(o, l)
		}
	}
	return o, nil
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
	stateRunning      = "running"
	stateStopped      = "stopped"
)

func AllTemplates() (vms Vms, e error) {
	return glob(TemplatesPath)
}

func AllWithTemplates() (vms Vms, e error) {
	return FindVms([]string{DefaultVMLocation, CloudVMLocation, TemplatesPath})
}

func AllWithIPsAndTags() (Vms, error) {
	vms, e := All()
	if e != nil {
		return nil, e
	}
	sort.Sort(vms)
	leases, e := AllLeases()
	if e != nil {
		return nil, e
	}
	arpInterfaces, err := LoadArpInterfaces()
	if err != nil {
		return nil, err
	}
	arpMap := map[string]*NetworkInterface{}
	for _, i := range arpInterfaces {
		arpMap[i.MAC.String()] = i
	}
	tags, e := LoadTags()
	if e != nil {
		return nil, e
	}

	tagsMap := map[string]string{}
	for _, t := range tags {
		tagsMap[t.Id()] = t.Value
	}
	for _, v := range vms {
		vmx, err := v.Vmx()
		if err != nil {
			return nil, err
		}

		mac := vmx.MacAddress
		if lease := leases.Lookup(mac); lease != nil {
			v.IP = lease.Ip
		} else {
			if m, err := net.ParseMAC(mac); err == nil {
				if i, ok := arpMap[m.String()]; ok && i.IP != nil {
					v.IP = i.IP.String()
				}
			}
		}
		v.Name = tagsMap[v.Id()+":Name"]
	}
	return vms, nil
}

func All() (Vms, error) {
	r, err := func() (map[string]struct{}, error) {
		l, err := Running()
		if err != nil {
			return nil, err
		}
		m := map[string]struct{}{}
		for _, l := range l {
			m[l] = struct{}{}
		}
		return m, nil
	}()
	if err != nil {
		return nil, err
	}
	vms, err := FindVms([]string{DefaultVMLocation, CloudVMLocation})
	if err != nil {
		return nil, err
	}
	for _, v := range vms {
		if _, ok := r[v.Path]; ok {
			v.State = stateRunning
		} else {
			v.State = stateStopped
		}
	}
	return vms, nil
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

	tags, e := LoadTags()
	if e != nil {
		return nil, e
	}
	tagsMap := map[string]Tags{}

	for _, tag := range tags {
		if tagsMap[tag.VmId] == nil {
			tagsMap[tag.VmId] = Tags{}
		}
		tagsMap[tag.VmId] = append(tagsMap[tag.VmId], tag)
	}
	m := map[string]bool{}
	for _, a := range raw {
		if _, ok := m[a.Path]; !ok {
			a.Tags = tagsMap[a.Id()]
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
	dbg.Printf("returning %d vms", len(vms))
	return vms, nil
}
