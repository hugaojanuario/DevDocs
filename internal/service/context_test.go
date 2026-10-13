package service

import (
	"context"
	"strings"
	"testing"

	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
)

type contextProvider struct{}

func (contextProvider) Search(context.Context, string, []string) ([]search.Result, error) {
	return []search.Result{{URL: "https://pandas.pydata.org/docs/reference/api/pandas.read_csv.html", Score: 100}, {URL: "https://pandas.pydata.org/docs/user_guide/io.html"}}, nil
}

type contextFetch struct{}

func (contextFetch) Fetch(_ context.Context, u string, _ int64) ([]byte, string, error) {
	if strings.Contains(u, "user_guide") {
		return []byte(`<main><h1>IO tools</h1><pre>import pandas as pd</pre></main>`), u, nil
	}
	return []byte(`<main><h1>pandas.read_csv</h1><p>Read CSV into a DataFrame using the documented reader.</p><pre>&gt;&gt;&gt; pd.read_csv("data.csv")
   Name Value
0  foo 1</pre></main>`), u, nil
}
func TestContextKeepsItsOwnSourceAndOutputIsNotCode(t *testing.T) {
	e := New(contextProvider{}, contextFetch{}, 0, 0)
	a, err := e.Search(context.Background(), model.SearchRequest{Query: "Como ler CSV com Pandas?"})
	if err != nil || len(a.Examples) != 1 || len(a.References) != 2 {
		t.Fatal(a, err)
	}
	ex := a.Examples[0]
	if ex.Code != `pd.read_csv("data.csv")` || ex.Setup != "import pandas as pd" || ex.SetupReferenceID != "source-2" || ex.Output != "   Name Value\n0  foo 1" || a.Summary == "" {
		t.Fatal(a)
	}
	if a.Evidence[1].Quote != ex.SourceCode || a.Evidence[2].ReferenceID != "source-2" {
		t.Fatal("lost provenance", a)
	}
}
