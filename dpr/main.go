package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var root = "/tmp/dpr"

type Resource struct {
	*http.Request
}

func (r *Resource) localPath() string {
	normalizedPath := root + r.Request.URL.Path
	if strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "index"
	}
	return normalizedPath
}

func (r *Resource) localHeadersPath() string {
	return r.localPath() + ".headers"
}

func (r *Resource) store() error {
	if e := r.storeHeaders(); e != nil {
		return e
	}
	if e := r.storeBody(); e != nil {
		return e
	}
	return nil
}

func (r *Resource) storeBody() error {
	f, e := r.createFile(r.localPath())
	if e != nil {
		return e
	}
	defer f.Close()
	i, e := io.Copy(f, r.Body)
	if e != nil {
		return e
	}
	log.Printf("stored %d byte", i)
	return nil
}

func (r *Resource) createFile(p string) (io.WriteCloser, error) {
	if e := os.MkdirAll(path.Dir(p), 0755); e != nil {
		return nil, e
	}
	log.Printf("creating file %s", p)
	return os.Create(p)
}

func (r *Resource) storeHeaders() error {
	f, e := r.createFile(r.localHeadersPath())
	if e != nil {
		return e
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(r.Header)
}

func (r *Resource) Write(w http.ResponseWriter) (int64, error) {
	log.Println("sending " + r.localPath())
	f, e := os.Open(r.localPath())
	if e != nil {
		return 0, e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return 0, e
	}
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	if strings.HasSuffix(r.Request.URL.String(), "/json") {
		stat, e := os.Stat(path.Dir(r.localPath()) + "/layer")
		if e == nil {
			w.Header().Add("X-Docker-Size", strconv.FormatInt(stat.Size(), 10))
		}
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(200)
	i, e := io.Copy(w, f)
	log.Printf("sent %d byte", i)
	return i, e
}

func (r *Resource) Exists() bool {
	return fileExists(r.localPath())
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

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + " http://" + r.Host + r.URL.String())
	defer r.Body.Close()

	normalizedPath := root + r.URL.Path
	if strings.HasSuffix(normalizedPath, "/") {
		normalizedPath += "index"
	}

	s := &Resource{r}
	w.Header().Set("X-Docker-Registry-Version", "0.0.1")
	w.Header().Add("X-Docker-Endpoints", r.Host)
	switch r.Method {
	case "PUT":
		e := s.store()
		if e != nil {
			log.Println(e.Error())
			http.Error(w, e.Error(), 500)
			return
		}
		w.Header().Add("WWW-Authenticate", `Token signature=123abc,repository="dynport/test",access=write`)
		w.Header().Add("X-Docker-Token", "token")
		w.Header().Add("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.String(), "/images") {
			w.WriteHeader(204)
		} else {
			w.WriteHeader(200)
		}
		return
	case "GET":
		if strings.HasSuffix(r.URL.String(), "/tags") {
			writeTags(s.localPath(), w)
		} else if strings.HasSuffix(r.URL.String(), "/ancestry") {
			p := root + path.Dir(r.URL.String())
			list := []string{path.Base(p)}
			for {
				img, e := loadImage(p + "/json")
				if e != nil {
					log.Print(e.Error())
					break
				}
				if img.Parent == "" {
					break
				}
				list = append(list, img.Parent)
				log.Println(list)
				p = path.Dir(p) + "/" + img.Parent
			}
			w.Header().Set("Content-Type", "application/json")
			if e := json.NewEncoder(w).Encode(list); e != nil {
				log.Print("ERROR: " + e.Error())
			}
		} else if s.Exists() {
			_, e := s.Write(w)
			if e != nil {
				log.Print("ERROR: " + e.Error())
			}
			return
		}
	}
	w.WriteHeader(404)
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

func main() {
	http.HandleFunc("/", ServeHTTP)
	addr := ":8088"
	log.Printf("starting dpr on %s", addr)
	e := http.ListenAndServe(addr, nil)
	if e != nil {
		log.Fatal(e.Error())
	}
}
