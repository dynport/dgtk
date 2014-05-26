package vmware

import (
	"encoding/json"
	"fmt"
	"os"
)

func UpdateTag(tag *Tag) error {
	tags := Tags{}
	e := tags.Load()
	if e != nil {
		return e
	}
	e = tags.Update(tag)
	if e != nil {
		return e
	}
	return tags.Store()
}

type Tag struct {
	VmId  string
	Key   string
	Value string
}

type Tags struct {
	tags []*Tag
}

func (tags *Tags) Tags() []*Tag {
	return tags.tags
}

func (list Tags) Len() int {
	return len(list.tags)
}

func (list Tags) Swap(a, b int) {
	list.tags[a], list.tags[b] = list.tags[b], list.tags[a]
}

func (list Tags) Less(a, b int) bool {
	return list.tags[a].Id() < list.tags[b].Id()
}

func (list *Tags) Update(tag *Tag) error {
	newTags := []*Tag{}
	handled := false
	for _, t := range list.tags {
		if t.VmId == tag.VmId && t.Key == tag.Key {
			handled = true
			if tag.Value != "" {
				newTags = append(newTags, tag)
			}
		} else {
			newTags = append(newTags, t)
		}
	}
	if !handled {
		newTags = append(newTags, tag)
	}
	list.tags = newTags
	return nil
}

func (tag *Tag) Id() string {
	return fmt.Sprintf("%s:%s", tag.VmId, tag.Key)
}

var tagsPath = os.ExpandEnv("$HOME/.vmware.tags")

func (list *Tags) Store() error {
	f, e := os.Create(tagsPath)
	if e != nil {
		return e
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(list.tags)
}

func (list *Tags) Load() error {
	f, e := os.Open(tagsPath)
	if e != nil {
		if os.IsNotExist(e) {
			return nil
		}
		return e
	}
	defer f.Close()
	list.tags = []*Tag{}
	e = json.NewDecoder(f).Decode(&list.tags)
	return e
}
