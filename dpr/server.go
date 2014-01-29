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
		return NewResource(server.DataRoot, r)
	}
}

func (server *Server) awsConfigured() bool {
	return server.AwsAccessKeyId != "" && server.AwsSecretAccessKey != ""
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
			p := server.DataRoot + path.Dir(r.URL.String())
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
				p = path.Dir(p) + "/" + img.Parent
			}
			w.Header().Set("Content-Type", "application/json")
			if e := json.NewEncoder(w).Encode(list); e != nil {
				log.Print("ERROR: " + e.Error())
			}
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
