package model

type SearchRequest struct {
	Technology       string   `json:"technology,omitempty"`
	Query            string   `json:"query"`
	Language         string   `json:"language"`
	PreferredSources []string `json:"preferred_sources"`
	AIEnabled        bool     `json:"ai_enabled"`
}
type Reference struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Domain     string `json:"domain"`
	Official   bool   `json:"official"`
	Technology string `json:"technology"`
	Version    string `json:"version"`
	Excerpt    string `json:"excerpt"`
}
type Explanation struct {
	Snippet     string `json:"snippet"`
	Description string `json:"description"`
	ReferenceID string `json:"reference_id"`
}
type Example struct {
	Language     string        `json:"language"`
	Code         string        `json:"code"`
	Explanations []Explanation `json:"explanations"`
	ReferenceID  string        `json:"reference_id"`
}
type Evidence struct {
	ReferenceID string `json:"reference_id"`
	Field       string `json:"field"`
	Quote       string `json:"quote"`
}
type Answer struct {
	Query        string      `json:"query"`
	Title        string      `json:"title"`
	Technology   string      `json:"technology"`
	Version      string      `json:"version"`
	Summary      string      `json:"summary"`
	Description  string      `json:"description"`
	Examples     []Example   `json:"examples"`
	Observations []string    `json:"observations"`
	References   []Reference `json:"references"`
	Evidence     []Evidence  `json:"evidence"`
	Mode         string      `json:"mode"`
	Status       string      `json:"status"`
	Quality      string      `json:"quality"`
	Language     string      `json:"language"`
}
