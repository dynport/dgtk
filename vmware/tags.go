package vmware

import (
	"encoding/json"
	"fmt"
	"os"
)

func UpdateTag(tag *Tag) error {
	tags, e := LoadTags()
	tags, e = tags.Update(tag)
	if e != nil {
		return e
	}
	return StoreTags(tags)
}

type Tag struct {
	VmId  string
	Key   string
	Value string
}

func LoadTags() (Tags, error) {
	f, e := os.Open(tagsPath)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	var tags Tags
	return tags, json.NewDecoder(f).Decode(&tags)
}

type Tags []*Tag

func (list Tags) Len() int {
	return len(list)
}

func (list Tags) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list Tags) Less(a, b int) bool {
	return list[a].Id() < list[b].Id()
}

func (list Tags) Update(tag *Tag) (Tags, error) {
	newTags := Tags{}
	handled := false
	for _, t := range list {
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
	return newTags, nil
}

func (tag *Tag) Id() string {
	return fmt.Sprintf("%s:%s", tag.VmId, tag.Key)
}

var tagsPath = os.ExpandEnv("$HOME/.vmware.tags")

func StoreTags(tags Tags) error {
	f, e := os.Create(tagsPath)
	if e != nil {
		return e
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tags)
}
