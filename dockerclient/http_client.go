package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func (dh *DockerHost) post(url string) (rsp *http.Response, e error) {
	rsp, e = dh.httpClient.Post(url, "", nil)
	if e == nil && !success(rsp) {
		e = fmt.Errorf("POST request to '%s' failed", url)
	}
	return rsp, e
}

func (dh *DockerHost) postWithBuffer(url string, buf *bytes.Buffer) (rsp *http.Response, e error) {
	rsp, e = dh.httpClient.Post(url, "", buf)
	if e == nil && !success(rsp) {
		e = fmt.Errorf("POST request to '%s' failed", url)
	}
	return rsp, e
}

func (dh *DockerHost) postJSON(url string, input interface{}, output interface{}) (content []byte, rsp *http.Response, e error) {
	buf := &bytes.Buffer{}
	if input != nil {
		json, e := json.Marshal(input)
		if e != nil {
			return nil, nil, e
		}
		buf.Write(json)
	}

	rsp, e = dh.httpClient.Post(url, "application/json", buf)
	if e != nil {
		return nil, rsp, e
	}
	defer rsp.Body.Close()
	content, e = ioutil.ReadAll(rsp.Body)

	if !success(rsp) {
		return content, rsp, fmt.Errorf("request failed: %d", rsp.StatusCode)
	}

	if output != nil {
		e = json.Unmarshal(content, output)
		if e != nil {
			e = fmt.Errorf("error unmarshalling: %s\n%s", e.Error(), string(content))
		}
	}

	return content, rsp, nil
}

func (dh *DockerHost) get(url string) (content []byte, rsp *http.Response, e error) {
	started := time.Now()

	rsp, e = dh.httpClient.Get(url)
	if e != nil {
		return nil, rsp, e
	}
	defer rsp.Body.Close()

	content, e = ioutil.ReadAll(rsp.Body)
	dh.Logger.Debugf("fetched %s in %.06f", url, time.Now().Sub(started).Seconds())
	return content, rsp, e
}

func (dh *DockerHost) getJSON(url string, i interface{}) (e error) {
	content, rsp, e := dh.get(url)
	if e != nil {
		return
	}
	if !success(rsp) {
		return fmt.Errorf("resource not found")
	}
	if i != nil {
		e = json.Unmarshal(content, i)
		if e != nil {
			return fmt.Errorf("error unmarshalling: %s\n%s", e.Error(), content)
		}
	}
	return nil
}

func success(rsp *http.Response) bool {
	return rsp.StatusCode >= 200 && rsp.StatusCode < 300
}
