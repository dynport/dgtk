package main

import (
	"github.com/dynport/dgtk/goproxy"
	"io"
	"log"
	"net/http"
)

func renderError(w http.ResponseWriter, e error) {
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, e.Error())
}

var ignoreNames = []string{"Packages.gz", "Release", "Release.gpg", "Sources.bz2", "Translation-en.bz2", "Translation-en_US", "Translation-en_US.bz2", "Translation-en_US.gz", "Translation-en_US.lzma", "Translation-en_US.xz", "Packages.bz2", "specs.4.8.gz", "prerelease_specs.4.8.gz"}

func main() {
	addr := ":1234"
	proxy := goproxy.Proxy{
		Address: addr,
	}
	log.Printf("starting goproxy with listen=%s with cached_dir=%q", proxy.Address, proxy.CacheDir())
	for _, name := range ignoreNames {
		proxy.Ignore(name)
	}
	e := proxy.Run()
	if e != nil {
		log.Fatal("ERROR: " + e.Error())
	}
}

func init() {
	log.SetFlags(0)
}
