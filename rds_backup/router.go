package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	r := cli.NewRouter()

	r.Register("snapshots/list", &list{}, "list all RDS snapshots")
	r.Register("snapshots/backup", &backup{}, "backup latest RDS snapshot")

	return r
}
