package main

import (
	"log"
)

type Delete struct {
	Name string `cli:"type=arg required=true"`
}

func (action *Delete) Run() error {
	log.Printf("deleting vm %s", action.Name)
	vm, e := findFirst(action.Name)
	if e != nil {
		return e
	}
	return vm.Delete()
}
