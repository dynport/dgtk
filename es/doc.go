package es

type Doc struct {
	Index  string      `json:"_index,omitempty"`
	Type   string      `json:"_type,omitempty"`
	Id     string      `json:"_id,omitempty"`
	Source interface{} `json:"-"`
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
