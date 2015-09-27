package main

import "github.com/dynport/dgtk/cli"

func router() *cli.Router {
	router := cli.NewRouter()

	router.Register("image/current", &currentImage{}, "get the current image")
	router.Register("env/current", &currentEnv{}, "get the current env")
	return router
}
