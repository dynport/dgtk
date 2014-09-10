package main

import (
	"fmt"
	"strconv"
)

type vmModify struct {
	Name   string `cli:"arg required"`
	Cpus   string `cli:"opt --cpus"`
	Memory string `cli:"opt --memory"`
}

func (r *vmModify) Run() error {
	vm, e := findFirst(r.Name)
	if e != nil {
		return e
	}
	if vm.Running() {
		return fmt.Errorf("vm must be running to be modified")
	}
	logger.Printf("modifying vm %q", vm.Path)
	modified := false
	if r.Cpus != "" {
		cpu, e := strconv.Atoi(r.Cpus)
		if e != nil {
			return e
		}
		e = vm.ModifyCpu(cpu)
		if e != nil {
			return e
		}
		modified = true
	}
	if r.Memory != "" {
		mem, e := strconv.Atoi(r.Memory)
		if e != nil {
			return e
		}
		e = vm.ModifyMemory(mem)
		return e
		modified = true
	}
	if !modified {
		return fmt.Errorf("either Memory or Cpus must be set to modify vm")
	}
	return nil
}
