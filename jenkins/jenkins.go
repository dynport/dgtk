package jenkins

import (
	"encoding/xml"
	"net/http"
)

type Jenkins struct {
	Address string
}

func (jenkins *Jenkins) Info() (info *Info, e error) {
	rsp, e := http.Get(jenkins.Address + "/api/xml?depth=1")
	if e != nil {
		return nil, e
	}
	defer rsp.Body.Close()
	dec := xml.NewDecoder(rsp.Body)
	info = &Info{}
	e = dec.Decode(info)
	return info, e
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
