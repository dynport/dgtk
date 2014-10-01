package dockerbuild

import (
	"log"
	"strings"

	"github.com/dynport/dgtk/dockerclient"
)

func (b *Build) callback(msg *dockerclient.JSONMessage) {
	if e := msg.Err(); e != nil {
		log.Printf("%s", e)
		return
	}
	if b.Verbose && msg.Stream != "" {
		log.Printf("%s", strings.TrimSpace(msg.Stream))
	}
}
