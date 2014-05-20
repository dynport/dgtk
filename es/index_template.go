package es

type IndexTemplate struct {
	Template string                 `json:"template,omitempty"`
	Settings *indexTemplateSettings `json:"settings,omitempty"`
}

type indexTemplateSettings struct {
	NumberOfShards   int                 `json:"number_of_shards"`
	NumberOfReplicas int                 `json:"mumber_of_replicas"`
	Mappings         map[string]*Mapping `json:"mappings,omitempty"`
}
