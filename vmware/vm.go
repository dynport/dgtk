package vmware

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Vm struct {
	Path string

	// cached values
	started time.Time
}

func (vm *Vm) Ip() (string, error) {
	leases, e := AllLeases()
	if e != nil {
		return "", e
	}
	vmx, e := vm.Vmx()
	if e != nil {
		return "", e
	}
	lease := leases.Lookup(vmx.MacAddress)
	if lease != nil {
		return lease.Ip, nil
	}
	return "", nil
}

func (vm *Vm) Running() bool {
	files, e := filepath.Glob(vm.dir() + "/*.vmem")
	if e != nil {
		return false
	}
	return len(files) > 0
}

func (vm *Vm) dir() string {
	return path.Dir(vm.Path)
}

func (vm *Vm) Name() string {
	return strings.TrimSuffix(path.Base(path.Dir(vm.Path)), ".vmwarevm")
}

func (vm *Vm) Vmx() (*Vmx, error) {
	vmx := &Vmx{}
	return vmx, vmx.Parse(vm.Path)
}

func (vm *Vm) GuestIPAddress() (string, error) {
	return GetGuestIPAddress(vm.Path)
}

func (vm *Vm) RestoreSnapshot(name string) error {
	return RestoreSnapshot(vm.Path, name)
}

func (vm *Vm) Stop() error {
	return Stop(vm.Path)
}

func (vm *Vm) Start() error {
	return Start(vm.Path, false)
}

func (vm *Vm) StartedAt() (started time.Time, e error) {
	if !vm.started.IsZero() {
		return vm.started, nil
	}
	f, e := os.Open(path.Dir(vm.Path) + "/vmware.log")
	if e != nil {
		return started, e
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "VMIOP: Init started") {
			parts := strings.Split(scanner.Text(), "|")
			if len(parts) > 1 {
				t, e := time.Parse("2006-01-02T15:04:05.999-07:00", parts[0])
				if e == nil {
					started = t
				}
			}
		}
	}
	vm.started = started
	return started, nil
}

func (vm *Vm) StartWithGui() error {
	return Start(vm.Path, true)
}

func (vm *Vm) Clone(dst string, opts *CloneOptions) (*Vm, error) {
	return Clone(vm.Path, dst, opts)
}

func (vm *Vm) Delete() error {
	return DeleteVM(vm.Path)
}

type Vms []*Vm

func (list Vms) Len() int {
	return len(list)
}

func (list Vms) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list Vms) Less(a, b int) bool {
	as, _ := list[a].StartedAt()
	bs, _ := list[b].StartedAt()
	return as.Before(bs)
}

func (list Vms) FindFirst(name string) *Vm {
	filtered := list.Search(name)
	if len(filtered) > 0 {
		return filtered[0]
	}
	return nil
}

func (list Vms) Search(name string) (vms Vms) {
	for _, vm := range list {
		if vm.Name() == name {
			vms = append(vms, vm)
		}
	}
	return vms
}
