package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	r := cli.NewRouter()

	r.Register("test", &testNotification{}, "test the hipchat notification mechanism")
	r.Register("notify", &sendNotification{}, "send notification if given command failed")

	return r
}
