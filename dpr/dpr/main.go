package main

import (
	"flag"
	"log"
	"os"

	"github.com/dynport/dgtk/dpr"
)

var (
	dataDir = flag.String("D", os.Getenv("HOME")+"/.dpr", "Location of data dir")
	addr    = flag.String("H", ":80", "Address to bind to")
	bucket  = flag.String("B", "", "S3 bucket to use for push")
	prefix  = flag.String("P", "", "S3 key prefix to use")
)

func main() {
	flag.Parse()
	server := &dpr.Server{
		DataRoot: *dataDir,
		Address:  *addr,
		Bucket:   *bucket,
		Prefix:   *prefix,
	}
	log.Printf("starting dpr on %s", server.Address)
	e := server.Run()
	if e != nil {
		log.Fatal("ERROR: " + e.Error())
	}
}
