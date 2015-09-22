package wunderproxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dynport/dgtk/wunderproxy/Godeps/_workspace/src/github.com/dynport/gocloud/aws/iam"
	"github.com/dynport/dgtk/wunderproxy/Godeps/_workspace/src/github.com/dynport/gocloud/aws/s3"
)

var ErrorEmptyHistory = errors.New("container history is empty")

type ContainerHistoryEvent struct {
	Revision   string
	DeployedAt time.Time
	DeployedBy string

	Hash    string
	history *ContainerHistory
}

// Load the event's container configuration from S3.
func (che *ContainerHistoryEvent) Load() (*LaunchConfig, error) {
	return LoadLaunchConfig(che.history.s3Client, che.history.s3Bucket, che.history.s3Prefix, che.Hash)
}

// A datastructure to manage the most recent containers that are/have been
// deployed. Including pointers to the launch configurations that were used
// for them.
type ContainerHistory struct {
	Events             []*ContainerHistoryEvent
	MaxSize            int
	s3Bucket, s3Prefix string
	s3Client           *s3.Client
}

// Load the container history from S3.
func LoadContainerHistory(s3c *s3.Client, bucket, prefix string) (*ContainerHistory, error) {
	ch := &ContainerHistory{s3Client: s3c, s3Bucket: bucket, s3Prefix: prefix}
	return ch, ch.load()
}

const maxHistoryEntries = 10

func (ch *ContainerHistory) load() error {
	resp, err := ch.s3Client.Get(ch.s3Bucket, ch.s3Prefix+"/history.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return ErrorEmptyHistory
	}

	err = json.NewDecoder(resp.Body).Decode(&ch.Events)
	if err != nil {
		return err
	}

	// If max size is not set in file, set it to 10.
	if ch.MaxSize == 0 {
		ch.MaxSize = maxHistoryEntries
	}

	for i := range ch.Events {
		ch.Events[i].history = ch
	}

	return nil
}

// Persist the container history to S3.
func (ch *ContainerHistory) Save() error {
	srcKey := ch.s3Prefix + "/container." + ch.Events[len(ch.Events)-1].Hash + ".json"
	err := ch.s3Client.Copy(ch.s3Bucket, srcKey, ch.s3Bucket, ch.s3Prefix+"/current.json")
	if err != nil {
		return err
	}

	if len(ch.Events) > maxHistoryEntries {
		ch.Events = ch.Events[len(ch.Events)-maxHistoryEntries:]
	}

	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(ch.Events)
	if err != nil {
		return err
	}

	return ch.s3Client.Put(ch.s3Bucket, ch.s3Prefix+"/history.json", buf.Bytes(), nil)
}

// Add a new entry to the history. This will add a new launch configuration and
// persist the launch configuration itself to S3. After a successful deployment
// (not part of this package) the changed history should be persisted, so that
// rollback and restart can rely on the current information.
func (ch *ContainerHistory) Add(lc *LaunchConfig) (*ContainerHistoryEvent, error) {
	err := lc.save(ch.s3Client, ch.s3Bucket, ch.s3Prefix)
	if err != nil {
		return nil, err
	}

	u, err := (&iam.Client{Client: ch.s3Client.Client}).GetUser("")
	if err != nil {
		return nil, err
	}

	ev := &ContainerHistoryEvent{
		history:    ch,
		Hash:       lc.hash,
		Revision:   lc.Revision,
		DeployedAt: time.Now(),
		DeployedBy: u.UserName,
	}

	ch.Events = append(ch.Events, ev)
	return ev, nil
}

// Rollback to the container in the given history event. The history is changed
// to only contain events up to the given one, i.e. all successors are removed.
// Please note an explicit history.Save call must be done to persist the
// changed history.
func (ch *ContainerHistory) RollbackTo(che *ContainerHistoryEvent) (*ContainerHistoryEvent, error) {
	i, err := func() (int, error) {
		for i := range ch.Events {
			if ch.Events[i].Hash == che.Hash {
				return i, nil
			}
		}
		return -1, fmt.Errorf("launch config %q not part of container history", che.Hash)
	}()
	if err != nil {
		return nil, err
	}

	ch.Events = ch.Events[0 : i+1]
	return ch.Events[i], nil
}
