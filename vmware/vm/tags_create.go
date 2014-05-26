package main

import "github.com/dynport/dgtk/vmware"

type TagsCreate struct {
	VmId  string `cli:"arg required"`
	Key   string `cli:"arg required"`
	Value string `cli:"arg required"`
}

func (r *TagsCreate) Run() error {
	tag := &vmware.Tag{VmId: r.VmId, Key: r.Key, Value: r.Value}
	return vmware.UpdateTag(tag)
}
