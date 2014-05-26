package main

import (
	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
)

type ListTemplates struct {
}

func (list *ListTemplates) Run() error {
	templates, e := vmware.AllTemplates()
	if e != nil {
		return e
	}
	table := gocli.NewTable()
	table.Add("Name", "Path")
	for _, t := range templates {
		table.Add(t.Id(), t.Path)
	}
	logger.Println(table)
	return nil
}
