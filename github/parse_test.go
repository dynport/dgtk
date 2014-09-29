package github

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/moovweb/gokogiri"
	. "github.com/smartystreets/goconvey/convey"
)

var logger = log.New(os.Stderr, "", 0)

type item struct {
	Type    string
	Name    string
	Payload []*payloadItem
}

const (
	stateBegin = iota
	stateEventName
	statePayload
	stateHookName
)

const (
	idxKey = iota
	idxType
	idxDescription
)

type payloadItem struct {
	Key         string
	Type        string
	Description string
}

func parseEventTypes(b []byte) ([]*item, error) {
	doc, e := gokogiri.ParseHtml(b)
	if e != nil {
		return nil, e
	}
	nodes, e := doc.Search("//div[@class='content']")
	if e != nil {
		return nil, e
	}
	if len(nodes) != 1 {
		return nil, fmt.Errorf("expected 1 content, found %d", len(nodes))
	}

	c := nodes[0]

	nodes, e = c.Search(".//*")
	if e != nil {
		return nil, e
	}
	i := &item{}
	items := []*item{}
	state := stateBegin
	for _, n := range nodes {
		switch n.Name() {
		case "h2":
			if i.Type != "" {
				items = append(items, i)
			}
			i = &item{Type: n.Content()}
		case "h3":
			switch id := n.Attr("id"); {
			case strings.HasPrefix(id, "event-name"):
				state = stateEventName
			case strings.HasPrefix(id, "payload"):
				state = statePayload
			case strings.HasPrefix(id, "hook-name"):
				state = stateHookName
			}
		case "table":
			if state == statePayload {
				trs, e := n.Search(".//tr")
				if e != nil {
					return nil, e
				}
				if len(trs) > 1 {
					payload := []*payloadItem{}
					for _, tr := range trs[1:] {
						tds, e := tr.Search(".//td")
						if e != nil {
							return nil, e
						}

						p := &payloadItem{}

						for i, td := range tds {
							c := strings.TrimSpace(td.Content())
							switch i {
							case idxKey:
								p.Key = c
							case idxType:
								p.Type = c
							case idxDescription:
								p.Description = c
							}

						}
						payload = append(payload, p)
					}
					if i != nil {
						i.Payload = payload
					}
				}
			}
		case "p":
			switch state {
			case stateEventName:
				i.Name = n.Content()
				state = stateBegin
			}
		}
	}
	return items, nil
}

func TestParse(t *testing.T) {
	Convey("Parsing", t, func() {
		b := mustRead(t, "event_types.html")
		So(b, ShouldNotBeNil)

		items, e := parseEventTypes(b)
		So(e, ShouldBeNil)
		So(items, ShouldNotBeNil)
		So(len(items), ShouldEqual, 22)

		item := items[0]
		So(item.Type, ShouldEqual, "CommitCommentEvent")
		So(item.Name, ShouldEqual, "commit_comment")

		So(item.Payload, ShouldNotBeNil)
		So(len(item.Payload), ShouldEqual, 1)

		So(item.Payload[0].Key, ShouldEqual, "comment")
		So(item.Payload[0].Type, ShouldEqual, "object")
		So(item.Payload[0].Description, ShouldEqual, "The comment itself.")

		item = items[10]
		So(item.Type, ShouldEqual, "GollumEvent")
		So(item.Name, ShouldEqual, "gollum")

		payload := item.Payload
		So(payload, ShouldNotBeNil)
		So(len(payload), ShouldEqual, 6)

		So(payload[1].Key, ShouldEqual, "pages[][page_name]")
		So(payload[1].Type, ShouldEqual, "string")
		So(payload[1].Description, ShouldEqual, "The name of the page.")

		types := map[string]int{}
		objects := map[string]int{}

		for _, i := range items {
			for _, p := range i.Payload {
				types[p.Type]++
				if p.Type == "object" {
					objects[p.Key]++
				}
			}
		}
		logger.Printf("types=%#v", types)
		logger.Printf("objects=%#v", objects)

	})
}

func mustRead(t *testing.T, name string) []byte {
	b, e := ioutil.ReadFile("fixtures/" + name)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
