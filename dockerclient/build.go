package dockerclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
)

type ProgressDetail struct {
	Current int64
	Total   int64
	Start   int64
}

type ErrorDetail struct {
	Code    int
	Message string
}

type BuildResponse []*JSONMessage

func (rsp BuildResponse) Last() *JSONMessage {
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

func ScanJson(r io.Reader, f func(b []byte) error) error {
	scanner := bufio.NewReader(r)
	buf := &bytes.Buffer{}
	for {
		b, e := scanner.ReadBytes('}')
		if e == io.EOF {
			break
		} else if e != nil {
			return e
		}
		e = f(b)
		if e != nil {
			return e
		}
		buf.Reset()
	}
	return nil
}

func (dh *DockerHost) handleBuildImageJson(r io.Reader, f func(s *JSONMessage)) (rsp BuildResponse, e error) {
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
		stream := &JSONMessage{}
		e = json.Unmarshal(buf.Bytes(), stream)
		if e != nil {
			if e.Error() == "unexpected end of JSON input" {
				continue
			}
			log.Printf("ERROR: %s => %s", e.Error(), buf.String())
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

func (dh *DockerHost) handleBuildImagePlain(r io.Reader, f func(s *JSONMessage)) (rsp BuildResponse, e error) {
	reader := bufio.NewReader(r)
	for {
		b, e := reader.ReadString('\n')
		if e == io.EOF {
			break
		} else if e != nil {
			return nil, e
		}
		s := &JSONMessage{Stream: string(b)}
		rsp = append(rsp, s)
		f(s)
	}
	return rsp, nil
}
