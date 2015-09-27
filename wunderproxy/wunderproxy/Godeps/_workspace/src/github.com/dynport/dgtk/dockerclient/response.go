package dockerclient

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type dockerStatusMessage struct {
	Error string `json:"error"`
}

// Required to split responses of docker. They come in as chunked HTTP messages that are joined automagically but then
// loose information on the single entities, each being a valid JSON string, and in total being invalid JSON. Split
// these messages and return them.
func splitDockerStatusMessages(body io.ReadCloser) (dSMs []*dockerStatusMessage, e error) {
	b, e := ioutil.ReadAll(body)
	if e != nil {
		return nil, e
	}

	dSMs = []*dockerStatusMessage{}

	last := 0
	totalLen := len(b)
	for i := 0; i < totalLen; i++ {
		if b[i] == '}' && (i+1 == totalLen || b[i+1] == '{') {
			dSM := &dockerStatusMessage{}
			if e = json.Unmarshal(b[last:i+1], dSM); e != nil {
				return nil, e
			}
			last = i + 1
			dSMs = append(dSMs, dSM)
		}
	}
	return dSMs, nil
}
