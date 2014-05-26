package main

import (
	"github.com/dynport/dgtk/cli"
	_ "github.com/dynport/gossh"
)

var router = cli.NewRouter()

func init() {
	router.Register("tags/create", &TagsCreate{}, "Create Tag")
	router.Register("tags/delete", &TagsDelete{}, "Delete Tag")
	router.Register("tags/list", &TagsList{}, "List Tags")
	router.Register("vms/clone", &Clone{}, "Clone VM")
	router.Register("vms/delete", &Delete{}, "Delete VM")
	router.Register("vms/list", &ListAction{}, "List VMs")
	router.Register("vms/start", &StartAction{}, "Start VM")
	router.Register("vms/stop", &StopAction{}, "Stop VM")
	router.Register("snapshots/list", &ListSnapshotsAction{}, "List Snapshots")
	router.Register("snapshots/restore", &ListSnapshotsAction{}, "Restore Snapshot")
	router.Register("snapshots/take", &ListSnapshotsAction{}, "Take Snapshot")
	router.Register("templates/list", &ListTemplates{}, "List Templates")
}

func main() {
	e := router.RunWithArgs()
	if e != nil {
		logger.Fatal("ERROR: " + e.Error())
	}
}
