package main

import (
	"fmt"

	"github.com/dynport/dgtk/vmware"
	"github.com/dynport/gocli"
)

type TagsList struct {
}

func (r *TagsList) Run() error {
	tags, e := vmware.LoadTags()
	if e != nil {
		return e
	}

	if tags.Len() == 0 {
		return nil
	}
	t := gocli.NewTable()
	for _, tag := range tags {
		t.Add(tag.VmId, tag.Key, tag.Value)
	}
	fmt.Println(t)
	return nil
}
