package es

type DynamicTemplateMapping struct {
	Type  string `json:"type,omitempty"`
	Index string `json:"index,omitempty"`
}

type DynamicTemplate struct {
	Match            string                  `json:"match,omitempty"`
	MatchMappingType string                  `json:"match_mapping_type,omitempty"`
	Mapping          *DynamicTemplateMapping `json:"mapping,omitempty"`
}

type DynamicTemplates []map[string]DynamicTemplate
