package search

import (
	"reflect"
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
