package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/codegangsta/martini"
	"github.com/dynport/dgtk/log"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "server":
			server()
		}
	}
}

func runCommandFromArgs() {
	if len(os.Args) > 2 {
		c := os.Args[2]
		args := []string{}
		if len(os.Args) > 3 {
			args = os.Args[3:]
		}
		log.Info("starting %s with %v", c, args)
		go func() {
			cmd := exec.Command(c, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()
		}()
	}
}

type Password struct {
	Name     string `json:"Name"`
	Login    string `json:"Login"`
	Password string `json:"Password"`
}

func mustRenderJSON(i interface{}, w http.ResponseWriter) (int, string) {
	b, e := json.Marshal(i)
	if e != nil {
		return http.StatusInternalServerError, e.Error()
	} else {
		w.Header().Set("Content-Type", "application/json")
		return http.StatusOK, string(b)
	}
}

var passwords = []*Password{}

func server() {
	m := martini.Classic()
	m.Use(martini.Static("assets"))
	m.Post("/passwords.json", func(r *http.Request, w http.ResponseWriter) (int, string) {
		defer r.Body.Close()
		b, e := ioutil.ReadAll(r.Body)
		if e != nil {
			return http.StatusInternalServerError, e.Error()
		}
		password := &Password{}
		e = json.Unmarshal(b, password)
		if e != nil {
			return http.StatusInternalServerError, e.Error()
		}
		passwords = append(passwords, password)
		return mustRenderJSON(password, w)
	})
	m.Get("/passwords.json", func(w http.ResponseWriter) (int, string) {
		return mustRenderJSON(passwords, w)
	})
	m.Get("/passwords/:Name.json", func(w http.ResponseWriter, params martini.Params) (int, string) {
		for _, p := range passwords {
			if p.Name == params["Name"] {
				return mustRenderJSON(p, w)
			}
		}
		return http.StatusNotFound, "Not Found"
	})
	m.Get("/.*", func(w http.ResponseWriter) (int, string) {
		b, e := ioutil.ReadFile("assets/layout.html")
		if e != nil {
			return http.StatusInternalServerError, e.Error()
		}
		w.Header().Set("Content-Type", "text/html")
		return http.StatusOK, string(b)
	})
	m.Run()
}
