package dockerclient

import (
	"fmt"
	"io"
	"strings"
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

type imageBuildResponseHandler struct {
	userHandler func(msg *JSONMessage)
	lastMessage *JSONMessage
}

func (ib *imageBuildResponseHandler) Run(r io.Reader) (string, error) {
	if e := handleJSONStream(r, ib.msgHandler); e != nil {
		return "", e
	}
	return ib.imageId()
}

func (ib *imageBuildResponseHandler) msgHandler(msg *JSONMessage) {
	if ib.userHandler != nil {
		ib.userHandler(msg)
	}
	ib.lastMessage = msg
}

func (ib *imageBuildResponseHandler) imageId() (string, error) {
	if ib.lastMessage == nil {
		return "", fmt.Errorf("failed to build")
	}

	if ib.lastMessage.Error != nil {
		return "", fmt.Errorf(ib.lastMessage.Error.Message)
	}

	prefix := "Successfully built "
	if strings.HasPrefix(ib.lastMessage.Stream, prefix) {
		return strings.TrimPrefix(ib.lastMessage.Stream, prefix), nil
	}

	return "", fmt.Errorf("unexpected message: %+v", ib.lastMessage)
}
