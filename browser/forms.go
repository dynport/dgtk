package browser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xml"
)

type Form struct {
	Action string
	Method string

	Inputs []*Input
}

func (f *Form) FillIn(name string, value string) error {
	for _, i := range f.Inputs {
		if i.Name == name {
			i.Value = value
			return nil
		}
	}
	return fmt.Errorf("no input with Name %q found", name)
}

type Input struct {
	Type    string
	Name    string
	Value   string
	Checked bool
}

func loadForms(baseUrl string, b []byte) ([]*Form, error) {
	u, e := url.Parse(baseUrl)
	if e != nil {
		return nil, e
	}
	doc, e := gokogiri.ParseHtml(b)
	if e != nil {
		return nil, e
	}
	formTags, e := doc.Search("//form")
	if e != nil {
		return nil, e
	}
	out := []*Form{}
	for _, node := range formTags {
		action := node.Attr("action")
		if !strings.HasPrefix(action, "http") {
			base := u.Scheme + "://" + u.Host
			if strings.HasPrefix(action, "/") {
				action = base + action
			} else {
				action = base + u.Path + "/" + strings.TrimPrefix(action, "/")
			}
		}
		f := &Form{Method: node.Attr("method"), Action: action}
		f.Inputs, e = loadInputs(node)
		if e != nil {
			return nil, e
		}
		out = append(out, f)
	}
	return out, nil
}

func loadInputs(doc xml.Node) ([]*Input, error) {
	nodes, e := doc.Search(".//input")
	if e != nil {
		return nil, e
	}
	out := []*Input{}
	for _, n := range nodes {
		i := &Input{
			Type:    n.Attr("type"),
			Name:    n.Attr("name"),
			Value:   n.Attr("value"),
			Checked: n.Attr("checked") == "checked",
		}
		out = append(out, i)
	}
	return out, nil

}
