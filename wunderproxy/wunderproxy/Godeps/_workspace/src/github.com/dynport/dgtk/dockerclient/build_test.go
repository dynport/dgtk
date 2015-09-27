package dockerclient

import (
	"bytes"
	"testing"
)

func TestParseBuildResponse(t *testing.T) {
	r := bytes.NewBufferString(newResponse)
	streams := BuildResponse{}
	if err := handleJSONStream(r, func(s *JSONMessage) {
		streams = append(streams, s)
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"len(streams)", 9, len(streams)},
		{"streams[0].Stream", streams[0].Stream, "Step 1 : FROM ubuntu\n"},
		{"streams.ImageID()", streams.ImageId(), "0f101a4836f6"},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}

const newResponse = `{"stream":"Step 1 : FROM ubuntu\n"}
{"stream":" ---\u003e 8dbd9e392a96\n"}
{"stream":"Step 2 : RUN apt-get update\n"}
{"stream":" ---\u003e Using cache\n"}
{"stream":" ---\u003e 30d9e1cb9bb8\n"}
{"stream":"Step 3 : RUN apt-get upgrade -y\n"}
{"stream":" ---\u003e Using cache\n"}
{"stream":" ---\u003e 0f101a4836f6\n"}
{"stream":"Successfully built 0f101a4836f6\n"}
`
