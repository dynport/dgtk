package es

type Hits struct {
	Total    int     `json:"total"`
	MaxScore float64 `json:"max_score"`
	Hits     []*Hit  `json:"hits"`
}
