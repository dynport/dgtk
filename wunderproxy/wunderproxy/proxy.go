package main

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/dynport/dgtk/wunderproxy"
)

type runAction struct {
	ProxyAddress string `cli:"opt --proxy default=0.0.0.0:80"`
	ApiAddress   string `cli:"opt --api default=0.0.0.0:8001"`
	RegistryPort int    `cli:"opt --registry-port default=8080"`
	ConfigFile   string `cli:"opt --cfg desc='Config file with the container env'"`
	S3Bucket     string `cli:"arg required"`
	S3Prefix     string `cli:"arg required"`
	AppName      string `cli:"arg required"`
}

func (r *runAction) Run() error {
	logger.Printf("running with %s and %s", r.ProxyAddress, r.ApiAddress)

	proxy := wunderproxy.NewProxy()

	cm, e := wunderproxy.NewContainerManager(r.S3Bucket, r.S3Prefix, r.AppName, proxy, r.ConfigFile, r.RegistryPort)
	if e != nil {
		return e
	}
	api := newWunderAPI(r.S3Bucket, r.S3Prefix, cm, proxy)

	wAPI := &http.Server{Addr: r.ApiAddress, Handler: api.Handler()}
	wProxy := &http.Server{Addr: r.ProxyAddress, Handler: proxy}

	wg := &sync.WaitGroup{}

	for _, s := range []*http.Server{wAPI, wProxy} {
		wg.Add(1)
		go startServer(s, wg)
	}
	wg.Wait()
	return nil
}

func startServer(s *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Printf("starting server on addr %q", s.Addr)

	err := func() error {
		switch {
		case strings.HasPrefix(s.Addr, "unix:"):
			ln, err := net.Listen("unix", strings.TrimPrefix(s.Addr, "unix:"))
			if err != nil {
				return err
			}
			return s.Serve(ln)
		default:
			return s.ListenAndServe()
		}
	}()

	if err != nil {
		logger.Printf("ERROR: %q", err)
	}
}
