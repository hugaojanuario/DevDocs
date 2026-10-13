package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

var ErrUnavailable = errors.New("search unavailable")
var ErrProvider = errors.New("search provider failed")

type Result struct {
	Title, URL, Snippet string
	Score               int
}
type ContentFetcher interface {
	Fetch(context.Context, string, int64) ([]byte, string, error)
}
type ProviderFetcher interface {
	Provider(context.Context, string, http.Header) ([]byte, error)
}
type SearchProvider interface {
	Search(context.Context, string, []string) ([]Result, error)
}
type Brave struct {
	Key      string
	Fetcher  ProviderFetcher
	Fallback SearchProvider
}

func (b *Brave) Search(ctx context.Context, q string, tech []string) ([]Result, error) {
	terms := strings.Join(Terms(q), " ") + " " + strings.Join(tech, " ")
	if terms == "" {
		terms = q
	}
	u := "https://api.search.brave.com/res/v1/web/search?q=" + url.QueryEscape(terms) + "&count=8"
	h := http.Header{}
	h.Set("X-Subscription-Token", b.Key)
	raw, e := b.Fetcher.Provider(ctx, u, h)
	if e != nil {
		if b.Fallback != nil {
			if v, e := b.Fallback.Search(ctx, q, tech); e == nil {
				return v, nil
			}
		}
		return nil, ErrProvider
	}
	var body struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	if e = json.Unmarshal(raw, &body); e != nil {
		return nil, ErrProvider
	}
	out := []Result{}
	for _, r := range body.Web.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Description})
	}
	if len(out) == 0 {
		return nil, ErrUnavailable
	}
	return Rank(out, tech), nil
}

type Official struct{ Fetcher ContentFetcher }
type sphinxIndex struct {
	Docnames   []string                   `json:"docnames"`
	Titles     []string                   `json:"titles"`
	Terms      map[string]json.RawMessage `json:"terms"`
	TitleTerms map[string]json.RawMessage `json:"titleterms"`
}

func (p *Official) Search(ctx context.Context, q string, tech []string) ([]Result, error) {
	if len(tech) == 0 {
		return nil, ErrUnavailable
	}
	out := []Result{}
	available := false
	for _, s := range Sources {
		if s.IndexURL == "" || !contains(tech, s.Technology) {
			continue
		}
		raw, _, e := p.Fetcher.Fetch(ctx, s.IndexURL, 8<<20)
		if e != nil {
			continue
		}
		idx, e := parseIndex(raw)
		if e != nil {
			continue
		}
		available = true
		scores := map[int]int{}
		for _, term := range Terms(q) {
			for _, key := range []string{term, strings.TrimSuffix(term, "s"), strings.TrimSuffix(term, "ing")} {
				for _, id := range indices(idx.Terms[key]) {
					scores[id] += 1
				}
				for _, id := range indices(idx.TitleTerms[key]) {
					scores[id] += 5
				}
			}
		}
		base, _ := url.Parse(s.IndexURL)
		for id, score := range scores {
			if score < 2 || id < 0 || id >= len(idx.Docnames) || id >= len(idx.Titles) {
				continue
			}
			rel := idx.Docnames[id] + ".html"
			if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
				continue
			}
			u := *base
			u.Path = path.Join(path.Dir(base.Path), rel)
			out = append(out, Result{Title: idx.Titles[id], URL: u.String(), Score: score + titleRelevance(idx.Titles[id], q, tech)})
		}
	}
	if !available || len(out) == 0 {
		return nil, ErrUnavailable
	}
	out = Rank(out, tech)
	if len(out) > 8 {
		out = out[:8]
	}
	return out, nil
}
func parseIndex(raw []byte) (sphinxIndex, error) {
	var idx sphinxIndex
	s := strings.TrimSpace(string(raw))
	if !strings.HasPrefix(s, "Search.setIndex(") {
		return idx, fmt.Errorf("unsupported official index")
	}
	s = strings.TrimPrefix(s, "Search.setIndex(")
	s = strings.TrimSuffix(strings.TrimSuffix(s, ";"), ")")
	e := json.Unmarshal([]byte(s), &idx)
	if len(idx.Docnames) == 0 {
		return idx, ErrUnavailable
	}
	return idx, e
}
func indices(raw json.RawMessage) []int {
	var n int
	if json.Unmarshal(raw, &n) == nil {
		return []int{n}
	}
	var a []int
	_ = json.Unmarshal(raw, &a)
	return a
}
func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
func Rank(results []Result, tech []string) []Result {
	for i := range results {
		s, ok := Classify(results[i].URL)
		if ok {
			results[i].Score += 100
			if contains(tech, s.Technology) {
				results[i].Score += 50
			}
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].URL < results[j].URL
		}
		return results[i].Score > results[j].Score
	})
	return results
}

// Function names such as read_csv should outrank pages that mention CSV only in their body.
func titleRelevance(title, query string, technologies []string) int {
	title = strings.ToLower(title)
	normalized := strings.NewReplacer("_", " ", ".", " ", "-", " ").Replace(title)
	terms := Terms(query)
	specific := []string{}
	score := 0
	for _, term := range terms {
		if contains(technologies, term) {
			continue
		}
		specific = append(specific, term)
		if strings.Contains(title, term) {
			score += 20
		}
	}
	for i := 0; i+1 < len(specific); i++ {
		if strings.Contains(normalized, specific[i]+" "+specific[i+1]) {
			score += 100
		}
	}
	return score
}
