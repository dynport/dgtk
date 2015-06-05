package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/dynport/gocli"
)

type nodesLS struct {
	Host string `cli:"opt -H default=127.0.0.1:9200"`
}

func (a *nodesLS) Run() error {
	h := a.Host
	if !strings.HasPrefix(h, "http") {
		h = "http://" + h
	}
	rsp, err := http.Get(h + "/_nodes/stats")
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	var r *nodeStatsResponse
	if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		return err
	}
	t := gocli.NewTable()
	list := NodeStats{}
	for _, n := range r.Nodes {
		list = append(list, n)
	}
	sort.Sort(list)
	t.Add("name", "host", "address", "load5", "cpu sys", "cpu user", "cpu idle", "mem used")
	for _, n := range list {
		load5 := ""
		if len(n.OS.LoadAverage) > 0 {
			load5 = fmt.Sprintf("%.2f", n.OS.LoadAverage[0])
		}
		t.Add(n.Name, n.Host, n.TransportAddress, load5, n.OS.CPU.Sys, n.OS.CPU.User, n.OS.CPU.Idle, n.OS.Mem.UsedPercent)
	}
	fmt.Println(t)
	return nil
}

type nodeStatsResponse struct {
	ClusterName string               `json:"cluster_name"`
	Nodes       map[string]*NodeStat `json:"nodes"`
}

type NodeStat struct {
	Timestamp        int64  `json:"timestamp"`
	Name             string `json:"name,omitempty"`
	Host             string `json:"host,omitempty"`
	HTTPAddress      string `json:"http_address"`
	TransportAddress string `json:"transport_address"`
	Version          string `json:"version,omitempty"`
	OS               struct {
		LoadAverage []float64 `json:"load_average"`
		CPU         struct {
			Sys   float64 `json:"sys,omitempty"`
			User  float64 `json:"user,omitempty"`
			Idle  float64 `json:"idle,omitempty"`
			Usage float64 `json:"usage,omitempty"`
		}
		Mem struct {
			UsedPercent float64 `json:"used_percent"`
			FreePercent float64 `json:"free_percent"`
		}
	} `json:"os"`
	Network struct {
		PrimaryInterface struct {
			Address string `json:"address,omitempty"`
		} `json:"primary_interface"`
	} `json:"network"`
}

type NodeStats []*NodeStat

func (list NodeStats) Len() int {
	return len(list)
}

func (list NodeStats) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list NodeStats) Less(a, b int) bool {
	return list[a].Name < list[b].Name
}
