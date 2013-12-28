package main

import (
	"fmt"
	"github.com/dynport/dgtk/vmware"
)

func findFirst(name string) (*vmware.Vm, error) {
	all, e := vmware.All()
	if e != nil {
		return nil, e
	}
	selected := all.Search(name)
	if len(selected) != 1 {
		return nil, fmt.Errorf("expected to find 1 got %v", selected)
	}
	return selected[0], nil
}
