package service

import (
	"context"
	"errors"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
	"sync/atomic"
	"testing"
	"time"
)

type provider struct {
	calls atomic.Int32
	err   error
}

func (p *provider) Search(context.Context, string, []string) ([]search.Result, error) {
	p.calls.Add(1)
	return []search.Result{{Title: "CSV", URL: "https://docs.python.org/3/library/csv.html"}}, p.err
}

type fetch struct{}

func (fetch) Fetch(context.Context, string, int64) ([]byte, string, error) {
	return []byte(`<html><main><h1>Read CSV</h1><p>Read CSV with the documented csv reader for data files.</p><pre>if True:
    read_csv("file.csv")</pre></main></html>`), "https://docs.python.org/3/library/csv.html", nil
}
func TestEvidenceCacheIsolationAndTTL(t *testing.T) {
	p := &provider{}
	e := New(p, fetch{}, 30*time.Millisecond, 1)
	req := model.SearchRequest{Query: "Python CSV"}
	a, err := e.Search(context.Background(), req)
	if err != nil || len(a.References) != 1 || len(a.Examples) != 1 || a.Evidence[0].Quote != a.Summary {
		t.Fatal(a, err)
	}
	a.Examples[0].Code = "mutated"
	b, _ := e.Search(context.Background(), req)
	if p.calls.Load() != 1 || b.Examples[0].Code == "mutated" {
		t.Fatal("cache alias")
	}
	time.Sleep(70 * time.Millisecond)
	e.mu.Lock()
	n := len(e.cache)
	e.mu.Unlock()
	if n != 0 {
		t.Fatal("expired data retained")
	}
	_, _ = e.Search(context.Background(), req)
	if p.calls.Load() != 2 {
		t.Fatal("expired entry served")
	}
	e.Clear()
	e.mu.Lock()
	n = len(e.cache)
	e.mu.Unlock()
	if n != 0 {
		t.Fatal("delete did not clear")
	}
}
func TestSearchFailureAndAI(t *testing.T) {
	p := &provider{err: search.ErrProvider}
	e := New(p, fetch{}, 0, 0)
	if _, err := e.Search(context.Background(), model.SearchRequest{Query: "Python CSV"}); !errors.Is(err, search.ErrProvider) {
		t.Fatal(err)
	}
	if _, err := e.Search(context.Background(), model.SearchRequest{AIEnabled: true}); !errors.Is(err, ErrAINotConfigured) {
		t.Fatal(err)
	}
}
