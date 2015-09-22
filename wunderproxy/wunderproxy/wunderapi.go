package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dynport/dgtk/wunderproxy"
)

func newWunderAPI(s3Bucket, s3Prefix string, cm *wunderproxy.ContainerManager, proxy *wunderproxy.Proxy) *wunderAPI {
	return &wunderAPI{cmanager: cm, proxy: proxy, s3bucket: s3Bucket, s3prefix: s3Prefix}
}

type wunderAPI struct {
	proxy    *wunderproxy.Proxy
	cmanager *wunderproxy.ContainerManager
	s3bucket string
	s3prefix string
}

func (a *wunderAPI) status(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		stats := map[string]interface{}{}

		err := a.proxy.Stats(stats)
		if err != nil {
			return err
		}
		err = a.cmanager.Stats(stats)
		if err != nil {
			return err
		}

		return json.NewEncoder(w).Encode(stats)
	}()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (a *wunderAPI) launchContainer(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		logger.Printf("launching container")
		config, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		logger.Printf("configuration url %q", config)
		containerId, port, err := a.cmanager.StartContainer(string(config))
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "started container %s with revision %s listening on port %d", containerId[:8], config[:8], port)
		return nil
	}()

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (a *wunderAPI) switchContainer(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		logger.Printf("switching container")
		config, e := ioutil.ReadAll(r.Body)
		if e != nil {
			return e
		}

		logger.Printf("configuration url %q", config)
		containerId, port, e := a.cmanager.SwitchContainer(string(config))
		if e != nil {
			return e
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "switched proxy to port %d of container %s", port, containerId[:8])
		return nil
	}()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
func (a *wunderAPI) maintenanceUp(w http.ResponseWriter, r *http.Request) {
	if e := a.proxy.MaintenanceUp(); e != nil {
		http.Error(w, e.Error(), 500)
	}
}

func (a *wunderAPI) maintenanceDown(w http.ResponseWriter, r *http.Request) {
	if e := a.proxy.MaintenanceDown(); e != nil {
		http.Error(w, e.Error(), 500)
	}
}

func (a *wunderAPI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", a.status)
	mux.HandleFunc("/launch", a.launchContainer)
	mux.HandleFunc("/switch", a.switchContainer)
	mux.HandleFunc("/maintenance/up", a.maintenanceUp)
	mux.HandleFunc("/maintenance/down", a.maintenanceDown)
	return mux
}
