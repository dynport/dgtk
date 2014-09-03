package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	r := cli.NewRouter()

	r.Register("snapshots/list", &listRDSSnapshots{}, "list all RDS snapshots")
	r.Register("snapshots/backup", &backupRDSSnapshot{}, "backup latest RDS snapshot")

	return r
}
