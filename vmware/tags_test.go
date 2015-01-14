package vmware

import (
	"sort"
	"testing"
)

func TestTags(t *testing.T) {
	var v, ex interface{}
	tags := Tags{}

	v = tags.Len()
	ex = 0
	if ex != v {
		t.Errorf("expected tags.Len() to be %#v, was %#v", ex, v)
	}

	tag := &Tag{VmId: "vm1", Key: "Name", Value: "This is the Name"}
	tags, err := tags.Update(tag)
	if err != nil {
		t.Fatal("error calling Update", err)
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"tags.Len()", 1, tags.Len()},
		{"tags[0].Key", "Name", tags[0].Key},
		{"tags[0].Value", "This is the Name", tags[0].Value},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}

	// updating an existing tag
	tag = &Tag{VmId: "vm1", Key: "Name", Value: "New Name"}
	tags, err = tags.Update(tag)
	if err != nil {
		t.Fatal("error calling update", err)
	}

	tests = []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"tags.Len()", 1, tags.Len()},
		{"tags[0].Key", "Name", tags[0].Key},
		{"tags[0].Value", "New Name", tags[0].Value},
	}

	// adding a new tag to the list
	tag = &Tag{VmId: "vm1", Key: "Enabled", Value: "true"}
	tags, err = tags.Update(tag)
	if err != nil {
		t.Fatal("error calling update", err)
	}

	sort.Sort(tags)

	tests = []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"tags.Len()", 2, tags.Len()},
		{"tags[0].Key", "Enabled", tags[0].Key},
		{"tags[0].Value", "true", tags[0].Value},
	}

	// removing a tag
	tag = &Tag{VmId: "vm1", Key: "Enabled", Value: ""}
	tags, err = tags.Update(tag)
	if err != nil {
		t.Fatal("error calling update", err)
	}
	v = 1
	ex = tags.Len()
	if ex != v {
		t.Errorf("expected tags.Len() to be %#v, was %#v", ex, v)
	}
}
