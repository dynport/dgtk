package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func handlePostResult(rsp *http.Response, err error) (*http.Response, error) {
	var e error
	if err == nil && !success(rsp) {
		content := ""
		if rsp.Body != nil {
			defer rsp.Body.Close()
			byteContent, e := ioutil.ReadAll(rsp.Body)
			if e == nil {
				content = string(byteContent)
			}
		}
		e = fmt.Errorf("POST request to %q failed (%d): %s", rsp.Request.URL.String(), rsp.StatusCode, content)
	}
	return rsp, e
}

func (dh *DockerHost) post(url string) (rsp *http.Response, e error) {
	return handlePostResult(dh.httpClient.Post(url, "", nil))
}

func (dh *DockerHost) postWithBuffer(url string, buf *bytes.Buffer) (rsp *http.Response, e error) {
	return handlePostResult(dh.httpClient.Post(url, "", buf))
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

	rsp, e = handlePostResult(dh.httpClient.Post(url, "application/json", buf))
	if e != nil {
		return nil, rsp, e
	}
	defer rsp.Body.Close()
	content, e = ioutil.ReadAll(rsp.Body)

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
