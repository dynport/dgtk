package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	dataDir = flag.String("D", "/data/dpr", "Location of data dir")
	addr    = flag.String("H", ":80", "Address to bind to")
	//awsAccessKeyId     = flag.String("aws-access-key-id", "", "AWS Access Key ID")
	//awsSecretAccessKey = flag.String("aws-secret-access-key", "", "AWS Secret Access Key")
)

func main() {
	flag.Parse()
	server := &Server{
		DataRoot: *dataDir,
		Address:  *addr,
	}
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

func writeTags(p string, w http.ResponseWriter) {
	files, e := filepath.Glob(p + "/*")
	if e != nil {
		log.Println("ERROR: " + e.Error())
		return
	}
	tags := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, ".headers") {
			continue
		}
		b, e := ioutil.ReadFile(f)
		if e != nil {
			continue
		}
		tags[path.Base(f)] = strings.Replace(string(b), `"`, "", -1)
	}
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
