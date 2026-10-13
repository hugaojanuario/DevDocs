package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
)

type helloFetch struct{ missing bool }

func (f helloFetch) Fetch(_ context.Context, u string, _ int64) ([]byte, string, error) {
	if strings.Contains(u, "functions.html") {
		return []byte(`<dl><dt id="print">print</dt><dd><p>Print objects to the text stream file.</p></dd></dl>`), u, nil
	}
	if f.missing {
		return []byte(`<pre>import asyncio</pre>`), u, nil
	}
	return []byte(`<pre>print("Hello World")</pre><pre>import asyncio</pre>`), u, nil
}

func TestBasicHelloIsShortAndGrounded(t *testing.T) {
	p := &provider{}
	e := New(p, helloFetch{}, 0, 0)
	a, err := e.Search(context.Background(), model.SearchRequest{Query: "Como fazer um hello world em Python?"})
	if err != nil || len(a.Examples) != 1 || a.Examples[0].Code != `print("Hello World")` || a.Description != "" || len(a.References) != 2 || p.calls.Load() != 0 {
		t.Fatal(a, err)
	}
	if a.Evidence[0].Quote != a.Summary || a.Evidence[1].Quote != a.Examples[0].Code {
		t.Fatal("missing literal evidence", a)
	}
	e = New(p, helloFetch{missing: true}, 0, 0)
	if _, err := e.Search(context.Background(), model.SearchRequest{Query: "hello world Python"}); !errors.Is(err, search.ErrUnavailable) {
		t.Fatal("invented fallback without source evidence", err)
	}
}

type manyExamples struct{}

func (manyExamples) Fetch(_ context.Context, u string, _ int64) ([]byte, string, error) {
	return []byte(`<main><h1>CSV</h1><p>Read CSV with a documented reader for data files.</p><p>Another long paragraph about CSV advanced options.</p><pre>read_csv("a.csv")</pre><pre>read_csv("b.csv")</pre><pre>read_csv("c.csv")</pre></main>`), u, nil
}

func TestDefaultAnswerKeepsOneMainExample(t *testing.T) {
	e := New(&provider{}, manyExamples{}, 0, 0)
	a, err := e.Search(context.Background(), model.SearchRequest{Query: "Python CSV"})
	if err != nil || len(a.Examples) != 1 || a.Description != "" || len(a.References) != 1 {
		t.Fatal(a, err)
	}
}
