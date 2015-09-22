package wunderproxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"runtime"
	"sync"
	"time"
)

type Proxy struct {
	Mutex        *sync.Mutex
	ConnStates   map[net.Conn]http.ConnState
	RequestStats struct {
		Total     int64
		TotalTime time.Duration
	}
	Address     string
	stateEvents chan *connStateEvent
	stats       chan *requestStat

	maintenanceMode  bool
	maintenancePath  string
	maintenanceCache []byte

	proxy *httputil.ReverseProxy
}

type connStateEvent struct {
	Conn  net.Conn
	State http.ConnState
}

type requestStat struct {
	TotalTime time.Duration
}

func NewProxy() *Proxy {
	p := &Proxy{Mutex: &sync.Mutex{}, ConnStates: map[net.Conn]http.ConnState{}}

	p.proxy = &httputil.ReverseProxy{Director: p.Director}

	p.stateEvents = make(chan *connStateEvent)
	p.stats = make(chan *requestStat)

	go func() {
		for stat := range p.stats {
			p.RequestStats.Total++
			p.RequestStats.TotalTime += stat.TotalTime
		}
	}()

	go func() {
		for event := range p.stateEvents {
			switch event.State {
			case http.StateClosed, http.StateHijacked:
				delete(p.ConnStates, event.Conn)
			default:
				p.ConnStates[event.Conn] = event.State
			}
		}
	}()

	return p
}

func (p *Proxy) Update(addr, maintenancePath string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	p.Address = addr
	p.maintenancePath = maintenancePath

	// Reset statistics
	p.RequestStats.Total = 0
	p.RequestStats.TotalTime = 0

	// Cache maintenance page
	p.cacheMaintenancePage()
}

func (p *Proxy) connState(con net.Conn, state http.ConnState) {
	p.stateEvents <- &connStateEvent{Conn: con, State: state}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()

	if p.Address == "" {
		http.NotFound(w, r)
		return
	}

	myW := &responseWriter{ResponseWriter: w}

	switch {
	case p.maintenanceMode:
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(500)
		_, _ = w.Write(p.maintenanceCache)
	default:
		p.proxy.ServeHTTP(myW, r)
	}

	totalTime := time.Since(started)
	p.stats <- &requestStat{TotalTime: totalTime}
	logger.Printf("method=%s url=%s status=%d total_time=%s", r.Method, r.URL.Path, myW.status, saneTotalTimePrinter(totalTime.Seconds()))
}

func saneTotalTimePrinter(tt float64) string {
	if tt < 1e-3 {
		return fmt.Sprintf("%.06fms", tt*1000)
	}
	return fmt.Sprintf("%.06fs", tt)
}

type responseWriter struct {
	http.ResponseWriter

	status int
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.ResponseWriter.WriteHeader(s)

	rw.status = s
}

func (p *Proxy) Director(req *http.Request) {
	req.URL.Scheme = "http"
	req.URL.Host = p.Address
}

func (p *Proxy) cacheMaintenancePage() {
	err := func() error {
		if p.Address == "" {
			return fmt.Errorf("no address set")
		}

		resp, err := http.Get("http://" + p.Address + "/" + p.maintenancePath)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		p.maintenanceCache, err = ioutil.ReadAll(resp.Body)
		return err
	}()
	if err != nil {
		logger.Printf("failed to cache maintenance page: %s", err)
		p.maintenanceCache = []byte("Page down, we're working on it.")
	}
}

func (p *Proxy) MaintenanceUp() error {
	if p.maintenanceCache == nil {
		p.cacheMaintenancePage()
	}
	p.maintenanceMode = true
	return nil
}

func (p *Proxy) MaintenanceDown() error {
	p.maintenanceMode = false
	return nil
}

func (p *Proxy) Stats(stats map[string]interface{}) error {
	memstat := &runtime.MemStats{}
	runtime.ReadMemStats(memstat)

	stats["HeapSize"] = memstat.HeapInuse
	stats["TotalTime"] = p.RequestStats.TotalTime.Seconds()
	stats["Requests"] = p.RequestStats.Total
	stats["MaintenanceMode"] = p.maintenanceMode

	states := map[string]int{}
	for _, v := range p.ConnStates {
		states[v.String()]++
	}
	stats["States"] = states

	return nil
}
