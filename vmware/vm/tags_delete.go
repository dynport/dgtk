package main

import "github.com/dynport/dgtk/vmware"

type TagsDelete struct {
	VmId string `cli:"arg required"`
	Key  string `cli:"arg required"`
}

func (r *TagsDelete) Run() error {
	tag := &vmware.Tag{VmId: r.VmId, Key: r.Key}
	return vmware.UpdateTag(tag)
}
