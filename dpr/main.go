package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
)

var (
	dataDir = flag.String("D", os.Getenv("HOME")+"/.dpr", "Location of data dir")
	addr    = flag.String("H", ":80", "Address to bind to")
)

func main() {
	flag.Parse()
	server := &Server{
		DataRoot:           *dataDir,
		Address:            *addr,
		AwsAccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AwsSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	log.Printf("aws: %v", server.awsConfigured())
	log.Printf("starting dpr on %s", server.Address)
	e := server.Run()
	if e != nil {
		log.Fatal("ERROR: " + e.Error())
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func writeTags(tags map[string]string, w http.ResponseWriter) {
	if e := json.NewEncoder(w).Encode(tags); e != nil {
		log.Println("ERROR: " + e.Error())
	}
}

func loadImage(p string) (*Image, error) {
	f, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	i := &Image{}
	e = json.NewDecoder(f).Decode(i)
	return i, e
}
