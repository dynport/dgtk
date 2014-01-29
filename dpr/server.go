package main

import (
	"encoding/json"
	"github.com/dynport/gocloud/aws/s3"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	DataRoot           string
	Address            string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.Address, s)
}

func (server *Server) newResource(r *http.Request) Resource {
	if server.awsConfigured() {
		client := s3.NewFromEnv()
		client.UseSsl = true
		client.CustomEndpointHost = "s3-eu-west-1.amazonaws.com"
		return &S3Resource{Request: r, Bucket: "de.1414.registry", Client: client}
	} else {
		return NewFileResource(server.DataRoot, r)
	}
}

func (server *Server) awsConfigured() bool {
	return server.AwsAccessKeyId != "" && server.AwsSecretAccessKey != ""
}

var ancestryCache = map[string]string{}
var ancestryMutex = &sync.Mutex{}

func lookupImageId(res Resource, imagePath string) (string, error) {
	imageId := path.Base(imagePath)
	ancestryMutex.Lock()
	defer ancestryMutex.Unlock()

	if parent, ok := ancestryCache[imageId]; ok {
		logger.Printf("found cached parent for %s (%s)", imageId, parent)
		return parent, nil
	}
	b, e := res.LoadResource(imagePath + "/json")
	if e != nil {
		return "", e
	}
	image := &Image{}
	e = json.Unmarshal(b, image)
	if e != nil {
		return "", e
	}
	ancestryCache[imageId] = image.Parent
	return image.Parent, nil
}

func loadAncestry(thePath string, res Resource) ([]string, error) {
	imagePath := path.Dir(thePath)
	imageRoot := path.Dir(imagePath)
	imageId := path.Base(imagePath)
	list := []string{imageId}
	for {
		parentId, e := lookupImageId(res, imagePath)
		if e != nil {
			return nil, e
		}
		if parentId == "" {
			return list, nil
		}
		list = append(list, parentId)
		imagePath = imageRoot + "/" + parentId
	}
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	log.Println(r.Method + " http://" + r.Host + r.URL.String())
	res := server.newResource(r)
	w.Header().Set("X-Docker-Registry-Version", "0.0.1")
	w.Header().Add("X-Docker-Endpoints", r.Host)
	switch r.Method {
	case "PUT":
		e := res.Store()
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
			tags, e := res.Tags()
			if e != nil {
				log.Printf(e.Error())
				w.WriteHeader(500)
				return
			}
			writeTags(tags, w)
			return
		} else if strings.HasSuffix(r.URL.String(), "/ancestry") {
			list, e := loadAncestry(r.URL.Path, res)
			if e != nil {
				http.Error(w, e.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			if e := json.NewEncoder(w).Encode(list); e != nil {
				log.Print("ERROR: " + e.Error())
			}
			return
		} else if res.Exists() {
			f, size, e := res.Open()
			if e != nil {
				http.Error(w, e.Error(), 500)
				return
			}
			if closer, ok := f.(io.Closer); ok {
				defer closer.Close()
			}
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

			if strings.HasSuffix(r.URL.String(), "/json") {
				w.Header().Set("Content-Type", "application/json")
				dockerSize, e := res.DockerSize()
				if e == nil {
					w.Header().Add("X-Docker-Size", strconv.FormatInt(dockerSize, 10))
				} else {
					logger.Print("ERROR: " + e.Error())
					http.Error(w, e.Error(), 500)
					return
				}
			}
			w.WriteHeader(200)
			_, e = io.Copy(w, f)
			if e != nil {
				log.Print("ERROR: " + e.Error())
			}
			return
		}
	}
	w.WriteHeader(404)
}
