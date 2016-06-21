package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/dynport/gocli"
)

type indexStats struct {
	Address string `cli:"opt -H default=http://127.0.0.1:9200"`
}

func (r *indexStats) Run() error {
	l := log.New(os.Stderr, "", 0)
	idx := &index{URL: normalizeIndexAddress(r.Address)}
	l.Printf("url=%s", idx.URL)

	stats, err := idx.indexStats()
	if err != nil {
		return fmt.Errorf("loading inexes: %s", err)
	}
	t := gocli.NewTable()
	t.Header("name", "count", "deleted")
	for _, stat := range stats {
		t.Add(stat.Name, stat.Total.Docs.Count, stat.Total.Docs.Deleted)
	}
	newline(l)
	l.Printf("indexes")
	sort.Sort(t)
	l.Printf("%s", t)
	newline(l)

	aliases, err := idx.aliases()
	if err != nil {
		return fmt.Errorf("loading aliases: %s", err)
	}
	t = gocli.NewTable()
	t.Header("alias", "indexes")
	for name, idxs := range aliases {
		t.Add(name, strings.Join(idxs, " "))
	}
	sort.Sort(t)
	l.Printf("aliases")
	l.Printf("%s", t)
	newline(l)

	health, err := idx.clusterHealth()
	if err != nil {
		return fmt.Errorf("loading cluster health: %s", err)
	}
	fmt.Printf("status=%s nodes=%d relocating=%d unassigned=%d\n", colorStatus(health.Status), health.NumberOfNodes, health.RelocatingShards, health.UnassignedShards)
	return nil
}

type Logger interface {
	Printf(string, ...interface{})
}

func newline(l Logger) {
	l.Printf("")
}

func colorStatus(status string) string {
	colorer := gocli.Green
	switch status {
	case "red":
		colorer = gocli.Red
	case "yellow":
		colorer = gocli.Yellow
	}
	return colorer(status)
}

type index struct {
	URL string
}

func (i *index) clusterHealth() (r *reponse, err error) {
	return r, loadJSON(i.URL+"/_cluster/health", &r)
}

type indexStat struct {
	Name  string
	Total struct {
		Docs struct {
			Count   int `json:"count"`
			Deleted int `json:"deleted"`
		} `json:"docs"`
	}
}

func (i *index) indexStats() (list []*indexStat, err error) {
	var rsp struct {
		Indices map[string]*indexStat `json:"indices"`
	}
	if err := loadJSON(i.URL+"/_stats", &rsp); err != nil {
		return nil, fmt.Errorf("loading index stats: %s", err)
	}
	for k, stat := range rsp.Indices {
		stat.Name = k
		list = append(list, stat)
	}
	return list, nil
}

type aliases map[string][]string

func (i *index) aliases() (cfg aliases, err error) {
	var rsp map[string]struct {
		Aliases map[string]interface{} `json:"aliases"`
	}
	if err := loadJSON(i.URL+"/_aliases", &rsp); err != nil {
		return nil, fmt.Errorf("loading alises: %s", err)
	}
	cfg = aliases{}
	for name, v := range rsp {
		for alias := range v.Aliases {
			cfg[alias] = append(cfg[alias], name)
		}
	}
	return cfg, nil
}

func loadJSON(url string, i interface{}) error {
	rsp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("requesting url %s: %s", url, err)
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}

	return json.NewDecoder(rsp.Body).Decode(&i)
}

type reponse struct {
	ClusterName         string `json:"cluster_name"`          // "elasticsearch",
	Status              string `json:"status"`                // "green",
	TimedOut            bool   `json:"timed_out"`             // false,
	NumberOfNodes       int    `json:"number_of_nodes"`       // 4,
	NumberOfDataNodes   int    `json:"number_of_data_nodes"`  // 4,
	ActivePrimaryShards int    `json:"active_primary_shards"` // 5,
	ActiveShards        int    `json:"active_shards"`         // 10,
	RelocatingShards    int    `json:"relocating_shards"`     // 0,
	InitializingShards  int    `json:"initializing_shards"`   // 0,
	UnassignedShards    int    `json:"unassigned_shards"`     // 0
}

func jq(in []byte) error {
	c := exec.Command("jq", ".")
	c.Stdin = bytes.NewReader(in)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func get(url string) ([]byte, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return nil, fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return ioutil.ReadAll(rsp.Body)
}
