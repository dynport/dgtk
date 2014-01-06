package dockerclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

type Stream struct {
	Stream      string `json:"stream"`
	Error       string `json:"error"`
	ErrorDetail string `json:"errorDetail"`
}

type BuildResponse []*Stream

func (rsp BuildResponse) Last() *Stream {
	if len(rsp) == 0 {
		return nil
	}
	return rsp[len(rsp)-1]
}

func (rsp BuildResponse) ImageId() string {
	last := rsp.Last()
	if last == nil {
		return ""
	}
	m := imageIdRegexp.FindStringSubmatch(last.Stream)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func (dh *DockerHost) handleBuildImageJson(r io.Reader, f func(s *Stream)) (rsp BuildResponse, e error) {
	scanner := bufio.NewReader(r)
	buf := &bytes.Buffer{}
	for {
		b, e := scanner.ReadBytes('}')
		if e == io.EOF {
			break
		} else if e != nil {
			return nil, e
		}
		buf.Write(b)
		stream := &Stream{}
		e = json.Unmarshal(buf.Bytes(), stream)
		if e != nil {
			if e.Error() == "unexpected end of JSON input" {
				continue
			}
			return nil, e
		}
		if f != nil {
			f(stream)
		}
		rsp = append(rsp, stream)
		buf.Reset()
	}
	return rsp, nil
}

func (dh *DockerHost) handleBuildImagePlain(r io.Reader, f func(s *Stream)) (rsp BuildResponse, e error) {
	reader := bufio.NewReader(r)
	for {
		b, e := reader.ReadString('\n')
		if e == io.EOF {
			break
		} else if e != nil {
			return nil, e
		}
		s := &Stream{Stream: string(b)}
		rsp = append(rsp, s)
		f(s)
	}
	return rsp, nil
}
