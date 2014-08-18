package main

import "github.com/dynport/dgtk/vmware"

type tagsCleanup struct {
}

func (r *tagsCleanup) Run() error {
	return vmware.CleanupTags()
}
