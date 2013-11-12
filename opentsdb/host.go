package opentsdb

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Host struct {
	Address string
}

func (host *Host) Load(path string) ([]*MetricValue, error) {
	url := host.Address + "/q?ascii&" + path
	log.Print("sending url " + url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Request to OpenTSDB failed with %s (%s)", resp.Status, string(b))
	}
	logger.Debug("Finished request to OpenTSDB")
	scanner := bufio.NewScanner(resp.Body)
	values := []*MetricValue{}
	for scanner.Scan() {
		line := scanner.Text()
		v := &MetricValue{}
		if e := v.Parse(line); e == nil {
			values = append(values, v)
		}
	}
	return values, nil
}
