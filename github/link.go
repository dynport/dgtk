package github

import "strings"

type Link struct {
	Last  string
	Next  string
	Prev  string
	First string
}

func ParseLink(s string) *Link {
	link := &Link{}
	fields := strings.Fields(s)
	last := ""
	for _, f := range fields {
		if strings.HasPrefix(f, "rel=") && strings.HasPrefix(last, "<") && strings.HasSuffix(last, ">;") {
			currentUrl := strings.TrimSuffix(strings.TrimPrefix(last, "<"), ">;")
			rel := strings.TrimSuffix(strings.Replace(strings.TrimPrefix(f, "rel="), `"`, "", -1), ",")
			switch rel {
			case "next":
				link.Next = currentUrl
			case "last":
				link.Last = currentUrl
			case "prev":
				link.Prev = currentUrl
			case "first":
				link.First = currentUrl
			}

		}
		last = f
	}
	return link
}
