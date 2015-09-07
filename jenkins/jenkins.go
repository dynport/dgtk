package jenkins

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	Address string

	// custom http client to be used
	Client *http.Client
}

func (j *Client) TriggerBuildWithParams(name string, values url.Values) (string, error) {
	//curl -u "dynport:horse battery staple" -d 'SPECS=spec&BRANCH=2.2.3&' https://phrasebuild.wunderscale.com/job/phrase-experimental/buildWithParameters -i
	u := j.Address + "/job/" + name + "/buildWithParameters"
	log.Printf("sending url=%s", u)
	rsp, err := j.client().PostForm(u, values)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.Status[0] != '2' {
		b, _ := ioutil.ReadAll(rsp.Body)
		return "", fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return rsp.Header.Get("Location"), nil
}

func (j *Client) QueueInfo() ([]byte, error) {
	return j.get("/queue/api/xml")
}

func (j *Client) client() *http.Client {
	if j.Client != nil {
		return j.Client
	}
	return http.DefaultClient
}

func (j *Client) Info() (info *Info, e error) {
	req, err := http.NewRequest("GET", j.Address+"/api/xml?depth=1", nil)
	if err != nil {
		return nil, err
	}
	rsp, err := j.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	dec := xml.NewDecoder(rsp.Body)
	info = &Info{}
	e = dec.Decode(info)
	return info, e
}

func (j *Client) JobConfig(name string) (*Config, error) {
	cfg := &Config{}
	b, err := j.get("/job/" + name + "/config.xml")
	if err != nil {
		return nil, err
	}
	cfg.RAW = string(b)
	err = xml.Unmarshal(b, &cfg)
	return cfg, err
}

func (j *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", j.Address+"/"+path, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := j.client().Do(req)
	if err != nil {
		return nil, err
	}
	b, _ := ioutil.ReadAll(rsp.Body)
	if rsp.Status[0] != '2' {
		return nil, fmt.Errorf("got status %s but expected 2x. body=%s", rsp.Status, string(b))
	}
	return b, nil
}

func (j *Client) loadXML(path string, i interface{}) (raw []byte, err error) {
	req, err := http.NewRequest("GET", j.Address+"/"+path, nil)
	if err != nil {
		return nil, err
	}
	rsp, err := j.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	buf := &bytes.Buffer{}
	dec := xml.NewDecoder(io.TeeReader(rsp.Body, buf))
	err = dec.Decode(&i)
	return buf.Bytes(), err
}

type Info struct {
	XMLName xml.Name `xml:"hudson"`
	Jobs    []*Job   `xml:"job"`
}

type Job struct {
	Name                string   `xml:"name"`
	Url                 string   `xml:"url"`
	Color               string   `xml:"color"`
	DisplayName         string   `xml:"display_name"`
	Buildable           bool     `xml:"buildable"`
	Builds              []*Build `xml:"build"`
	FirstBuild          *Build   `xml:"firstBuild"`
	LastBuild           *Build   `xml:"lastBuild"`
	LastSuccessfulBuild *Build   `xml:"lastSuccessfulBuild"`
	LastStableBuild     *Build   `xml:"lastStableBuild"`
	LastCompletedBuild  *Build   `xml:"lastCompletedBuild"`
}

type Build struct {
	Number string `xml:"number"`
	Url    string `xml:"url"`
}
