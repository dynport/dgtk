package vmware

import (
	"path"
	"path/filepath"
	"strings"
)

type Vm struct {
	Path string
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

func (list Vms) Search(name string) (vms Vms) {
	for _, vm := range list {
		if vm.Name() == name {
			vms = append(vms, vm)
		}
	}
	return vms
}
