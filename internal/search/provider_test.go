package search

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRegistryAndRanking(t *testing.T) {
	for _, u := range []string{"https://docs.python.org.evil.test/3/", "https://postgresql.org/download/", "https://evil.test/docs.python.org/"} {
		if _, ok := Classify(u); ok {
			t.Fatal(u)
		}
	}
	r := Rank([]Result{{URL: "https://example.org/tutorial"}, {URL: "https://pandas.pydata.org/docs/reference.html"}}, []string{"pandas"})
	if r[0].URL != "https://pandas.pydata.org/docs/reference.html" {
		t.Fatal(r)
	}
	if !reflect.DeepEqual(Detect("Como ler CSV usando Pandas?"), []string{"pandas"}) {
		t.Fatal(Detect("Como ler CSV usando Pandas?"))
	}
}
func TestOfficialIndex(t *testing.T) {
	idx, e := parseIndex([]byte(`Search.setIndex({"docnames":["reference/read"],"titles":["Read"],"terms":{"csv":[0]},"titleterms":{"read":0}});`))
	if e != nil || len(idx.Docnames) != 1 || indices(idx.Terms["csv"])[0] != 0 {
		t.Fatal(idx, e)
	}
	if _, e = parseIndex([]byte(`alert("fake")`)); e == nil {
		t.Fatal("accepted invalid index")
	}
}

type fakeFetch struct {
	err error
	raw []byte
}

func (f fakeFetch) Fetch(context.Context, string, int64) ([]byte, string, error) {
	return f.raw, "https://docs.python.org/3/searchindex.js", f.err
}
func (f fakeFetch) Provider(context.Context, string, http.Header) ([]byte, error) {
	return f.raw, f.err
}
func TestProviderFailureAndOfficialFallback(t *testing.T) {
	idx := []byte(`Search.setIndex({"docnames":["library/csv"],"titles":["CSV"],"terms":{"csv":0},"titleterms":{"csv":0}});`)
	official := &Official{Fetcher: fakeFetch{raw: idx}}
	r, e := official.Search(context.Background(), "Python CSV", []string{"python"})
	if e != nil || len(r) != 1 || r[0].URL != "https://docs.python.org/3/library/csv.html" {
		t.Fatal(r, e)
	}
	b := &Brave{Fetcher: fakeFetch{err: errors.New("upstream down")}, Fallback: official}
	if r, e = b.Search(context.Background(), "Python CSV", []string{"python"}); e != nil || len(r) != 1 {
		t.Fatal("fallback failed")
	}
	b.Fallback = nil
	if _, e = b.Search(context.Background(), "Python CSV", nil); !errors.Is(e, ErrProvider) {
		t.Fatal(e)
	}
	if _, e = official.Search(context.Background(), "unknown", nil); !errors.Is(e, ErrUnavailable) {
		t.Fatal(e)
	}
}

func TestFunctionTitleRanksBeforeGeneralMentions(t *testing.T) {
	if titleRelevance("pandas.read_csv", "Como ler um CSV com Pandas?", []string{"pandas"}) <= titleRelevance("Comparison with spreadsheets", "Como ler um CSV com Pandas?", []string{"pandas"}) {
		t.Fatal("function relevance lost")
	}
}

func TestBasicQuestionPrefersIntroductionOverAdministration(t *testing.T) {
	idx := []byte(`Search.setIndex({"docnames":["administration/dag-serialization","core-concepts/dags"],"titles":["Dag Serialization","Dags"],"terms":{"create":[0,1],"dag":[0,1],"airflow":0},"titleterms":{"dag":[0,1]}});`)
	p := &Official{Fetcher: fakeFetch{raw: idx}}
	results, err := p.Search(context.Background(), "Como criar um DAG no Airflow?", []string{"airflow"})
	if err != nil || len(results) != 2 || !strings.Contains(results[0].URL, "core-concepts/dags") {
		t.Fatal(results, err)
	}
	if got := TopicTerms("Como ler CSV com Python e Pandas?", []string{"python", "pandas"}); !reflect.DeepEqual(got, []string{"read", "csv"}) {
		t.Fatal(got)
	}
}
