package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hugojanuario/devdocs/internal/extraction"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Optional future synthesis must consume only collected evidence. No AI provider is enabled in V1.
type AnswerProvider interface {
	Answer(context.Context, model.Answer) (model.Answer, error)
}
type Searcher interface {
	Search(context.Context, model.SearchRequest) (model.Answer, error)
}
type entry struct {
	raw     []byte
	expires time.Time
}
type Engine struct {
	Provider   search.SearchProvider
	Fetcher    search.ContentFetcher
	TTL        time.Duration
	Capacity   int
	mu         sync.Mutex
	cache      map[[32]byte]entry
	generation uint64
	slots      chan struct{}
}

func New(provider search.SearchProvider, fetcher search.ContentFetcher, ttl time.Duration, capacity int) *Engine {
	return &Engine{Provider: provider, Fetcher: fetcher, TTL: ttl, Capacity: capacity, cache: map[[32]byte]entry{}, slots: make(chan struct{}, 4)}
}
func (e *Engine) Search(ctx context.Context, r model.SearchRequest) (model.Answer, error) {
	if r.AIEnabled {
		return model.Answer{}, search.ErrUnavailable
	}
	r.Query = strings.Join(strings.Fields(r.Query), " ")
	if r.Language == "" {
		r.Language = "pt-BR"
	}
	raw, _ := json.Marshal(r)
	key := sha256.Sum256(raw)
	e.mu.Lock()
	now := time.Now()
	for k, v := range e.cache {
		if now.After(v.expires) {
			delete(e.cache, k)
		}
	}
	generation := e.generation
	cached, ok := e.cache[key]
	e.mu.Unlock()
	if ok {
		var a model.Answer
		_ = json.Unmarshal(cached.raw, &a)
		return a, nil
	}
	select {
	case e.slots <- struct{}{}:
		defer func() { <-e.slots }()
	default:
		return model.Answer{}, search.ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	technologies := search.Detect(r.Query)
	if r.Technology != "" {
		technologies = []string{r.Technology}
	}
	results, err := e.Provider.Search(ctx, r.Query, technologies)
	if err != nil {
		return model.Answer{}, err
	}
	a := model.Answer{Query: r.Query, Language: r.Language, Examples: []model.Example{}, References: []model.Reference{}, Evidence: []model.Evidence{}, Observations: []string{}, Mode: "deterministic", Status: "completed", Quality: "extracted"}
	type fetched struct {
		result   search.Result
		doc      extraction.Document
		final    string
		source   search.Source
		official bool
		err      error
	}
	work := []search.Result{}
	seen := map[string]bool{}
	for _, res := range results {
		s, official := search.Classify(res.URL)
		_ = s
		if len(r.PreferredSources) == 1 && ((r.PreferredSources[0] == "official" && !official) || (r.PreferredSources[0] == "complementary" && official)) {
			continue
		}
		u, err := url.Parse(res.URL)
		if err != nil {
			continue
		}
		u.Fragment = ""
		if seen[u.String()] {
			continue
		}
		seen[u.String()] = true
		res.URL = u.String()
		work = append(work, res)
		if len(work) == 4 {
			break
		}
	}
	docs := make([]fetched, len(work))
	var wg sync.WaitGroup
	for i, res := range work {
		wg.Add(1)
		go func(i int, res search.Result) {
			defer wg.Done()
			raw, final, err := e.Fetcher.Fetch(ctx, res.URL, 2<<20)
			d := fetched{result: res, final: final, err: err}
			if err == nil {
				d.source, d.official = search.Classify(final)
				if len(r.PreferredSources) == 1 && r.PreferredSources[0] == "official" && !d.official {
					d.err = errors.New("redirect left official source")
				} else {
					d.doc, d.err = extraction.Extract(raw, search.Terms(r.Query), fmt.Sprintf("source-%d", i+1), d.source.Technology)
				}
			}
			docs[i] = d
		}(i, res)
	}
	wg.Wait()
	for i, d := range docs {
		if d.err != nil {
			a.Status = "partial"
			continue
		}
		if d.doc.Summary == "" && len(d.doc.Examples) == 0 {
			a.Status = "partial"
			continue
		}
		id := fmt.Sprintf("source-%d", i+1)
		u, _ := url.Parse(d.final)
		version := d.doc.Version
		if version == "" {
			version = strings.Trim(search.Version(d.final), "/")
		}
		title := d.doc.Title
		if title == "" {
			title = d.result.Title
		}
		a.References = append(a.References, model.Reference{ID: id, Title: title, URL: d.final, Domain: u.Hostname(), Official: d.official, Technology: d.source.Technology, Version: version, Excerpt: d.doc.Summary})
		if a.Title == "" {
			a.Title = title
			a.Technology = d.source.Technology
			a.Version = version
			a.Summary = d.doc.Summary
			a.Description = d.doc.Description
			if a.Summary != "" {
				a.Evidence = append(a.Evidence, model.Evidence{ReferenceID: id, Field: "summary", Quote: a.Summary})
			}
			if a.Description != "" {
				a.Evidence = append(a.Evidence, model.Evidence{ReferenceID: id, Field: "description", Quote: a.Description})
			}
		}
		if len(a.Examples) < 6 {
			for _, ex := range d.doc.Examples {
				if len(a.Examples) >= 6 {
					break
				}
				a.Examples = append(a.Examples, ex)
				a.Evidence = append(a.Evidence, model.Evidence{ReferenceID: id, Field: "example", Quote: ex.Code})
			}
		}
	}
	if len(a.References) == 0 {
		return model.Answer{}, search.ErrUnavailable
	}
	if a.Title == "" {
		a.Title = r.Query
	}
	if a.Summary == "" || len(a.Examples) == 0 {
		a.Quality = "limited"
		a.Status = "partial"
	}
	if r.Language == "en" {
		a.Observations = append(a.Observations, "Verbatim documentation excerpts. Source language and indentation are preserved. Automatic line-by-line explanation is not guaranteed.")
	} else {
		a.Observations = append(a.Observations, "Trechos literais da documentação. Idioma original e indentação preservados. Explicação automática por linha não é garantida.")
	}
	requestedVersion := search.Version(r.Query)
	if requestedVersion != "" && (a.Version == "" || !strings.HasPrefix(a.Version, requestedVersion)) {
		if r.Language == "en" {
			a.Observations = append(a.Observations, "The requested version could not be verified in the collected sources. Check each reference's version.")
		} else {
			a.Observations = append(a.Observations, "A versão solicitada não foi confirmada nas fontes coletadas. Confira a versão de cada referência.")
		}
	}
	raw, _ = json.Marshal(a)
	if e.TTL > 0 && e.Capacity > 0 && len(raw) <= 256<<10 {
		e.mu.Lock()
		if e.generation != generation {
			e.mu.Unlock()
			return a, nil
		}
		if len(e.cache) >= e.Capacity {
			for k := range e.cache {
				delete(e.cache, k)
				break
			}
		}
		expires := time.Now().Add(e.TTL)
		e.cache[key] = entry{raw: raw, expires: expires}
		e.mu.Unlock()
		time.AfterFunc(e.TTL, func() {
			e.mu.Lock()
			defer e.mu.Unlock()
			if v, ok := e.cache[key]; ok && v.expires.Equal(expires) {
				delete(e.cache, key)
			}
		})
	}
	return a, nil
}

// Clear removes transient search data, including after a saved query is deleted.
func (e *Engine) Clear() { e.mu.Lock(); clear(e.cache); e.generation++; e.mu.Unlock() }
