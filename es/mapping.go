package es

type IndexMappings map[string]*IndexMapping

type IndexMapping struct {
	Properties *IndexMappingProperties
}

type IndexMappingProperties map[string]IndexMappingProperty

type IndexMappingProperty struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

type Mapping map[string]IndexMappings

func (mapping Mapping) IndexNames() []string {
	names := []string{}
	for index, _ := range mapping {
		names = append(names, index)
	}
	return names
}
