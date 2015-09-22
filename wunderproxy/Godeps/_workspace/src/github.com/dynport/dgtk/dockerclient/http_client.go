package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
	return dh.postWithReader(url, nil)
}

func (dh *DockerHost) postWithReader(url string, r io.Reader) (rsp *http.Response, e error) {
	return dh.postWithContentType(url, "", r)
}

func (dh *DockerHost) postWithContentType(url, contentType string, r io.Reader) (rsp *http.Response, e error) {
	return handlePostResult(dh.httpClient.Post(url, contentType, r))
}

func (dh *DockerHost) postJSON(url string, input interface{}, output interface{}) (rsp *http.Response, e error) {
	buf := &bytes.Buffer{}
	if e = json.NewEncoder(buf).Encode(input); e != nil {
		return nil, e
	}

	rsp, e = dh.postWithContentType(url, "application/json", buf)
	if e != nil {
		return rsp, e
	}
	defer rsp.Body.Close()

	var content []byte
	if content, e = ioutil.ReadAll(rsp.Body); e != nil {
		return nil, fmt.Errorf("failed reading content: %s", e)
	}

	if output != nil {
		if e = json.Unmarshal(content, output); e != nil {
			return nil, fmt.Errorf("error unmarshalling: %s\n%s", e.Error(), string(content))
		}
	}

	return rsp, nil
}

func (dh *DockerHost) get(url string) (content []byte, rsp *http.Response, e error) {
	rsp, e = dh.httpClient.Get(url)
	if e != nil {
		return nil, rsp, e
	}
	defer rsp.Body.Close()

	content, e = ioutil.ReadAll(rsp.Body)
	return content, rsp, e
}

var ErrorNotFound = fmt.Errorf("resource not found")

func (dh *DockerHost) getJSON(url string, i interface{}) (e error) {
	content, rsp, e := dh.get(url)
	if e != nil {
		return
	}
	if rsp.StatusCode == http.StatusNotFound {
		return ErrorNotFound
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
