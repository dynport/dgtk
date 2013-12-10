package es

type Doc struct {
	Index  string
	Type   string
	Id     string
	Source interface{}
}

func (doc *Doc) IndexAttributes() map[string]string {
	atts := map[string]string{}
	if doc.Index != "" {
		atts["_index"] = doc.Index
	}
	if doc.Type != "" {
		atts["_type"] = doc.Type
	}
	if doc.Id != "" {
		atts["_id"] = doc.Id
	}
	return atts
}
